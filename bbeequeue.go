// Copyright 2025 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package bbeequeue

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
)

// New returns a new channel for receiving T values from the specified eBPF
// ringbuffer map until the passed context gets done. When the context is done,
// the channel will close automatically.
//
// New optionally takes configuration options (that's why they are called
// options ... “my name is Option, Optional Option”):
//
//   - [WithSize] configures the channel buffer size; it defaults to 0 in which
//     case the returned channel is unbuffered. New will return an error when
//     trying to configure a negative size.
//   - [WithErrorChannel] configures a caller-supplied channel that will receive
//     any errors occurring when reading and unmarshalling data from the ringbuffer.
//     Please note that this channel will never be closed automatically.
func New[T any](ctx context.Context, rbmap *ebpf.Map, opts ...Option) (<-chan T, error) {
	var options options
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	rring, err := ringbuf.NewReader(rbmap)
	if err != nil {
		return nil, err
	}

	// Monitor the passed context in the background: when the context gets done
	// we close the ringbuffer reader in order to unblock the message pump below
	// with an ErrClosed.
	go func() {
		<-ctx.Done()
		_ = rring.Close()
	}()

	// We're now kicking off another background go routine that acts as the
	// event pump: it waits for a new record to arrive in the ringbuffer,
	// unmarshals it and then stuffs into the event channel. Rinse and repeat.
	//
	// Now, as cilium/ebpf only supports blocking (or otherwise time-boxed)
	// reads from an eBPF ringbuffer but without any context control, we rely on
	// the above context monitoring go routine to kick us hard by closing our
	// reader while we're blocked waiting for new records. Then we call it a
	// day.
	ch := make(chan T, options.size)
	go func() {
		defer close(ch)

		var event T
		for {
			record, err := rring.Read()
			if err != nil {
				if errors.Is(err, ringbuf.ErrClosed) {
					return
				}
				if options.errch != nil {
					options.errch <- err
				}
				continue
			}
			if err := binary.Read(
				bytes.NewBuffer(record.RawSample), binary.NativeEndian, &event); err != nil {
				if options.errch != nil {
					options.errch <- err
				}
				continue
			}
			// Ensure that while we're stuck waiting to push the received record
			// into the channel, we additionally monitor the context: if it
			// get's done, then we abort our attempt to send the record down the
			// channel. In this case, the other go routine above also monitoring
			// the context will make sure to close the ringbuffer reader,
			// releasing associated resources.
			select {
			case ch <- event:
			// That's fine.
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}
