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

func New[T any](ctx context.Context, m *ebpf.Map, size int) (chan T, error) {
	rring, err := ringbuf.NewReader(m)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = rring.Close()
	}()

	ch := make(chan T, size)
	go func() {
		defer close(ch)

		var event T
		for {
			record, err := rring.Read()
			if err != nil && errors.Is(err, ringbuf.ErrClosed) {
				return
			}
			if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.NativeEndian, &event); err != nil {
				continue
			}
			ch <- event
		}
	}()
	return ch, nil
}
