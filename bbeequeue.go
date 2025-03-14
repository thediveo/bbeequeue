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
