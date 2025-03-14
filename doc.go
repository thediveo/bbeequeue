/*
Package bbeequeue supports receiving data from type BPF_MAP_TYPE_RINGBUF eBPF
maps using idiomatic Go channels. This package does not attempt to be a
jack-of-all-trades implementation, but instead to cover the basic use case of
providing channel access to eBPF ringbuffer maps.

# Usage

[New] returns a new channel producing T values until the passed context gets
done. The default is an unbuffered channel, specify [WithSize] in the call to
New in order to create a buffered channel of the configured size.

	import "github.com/thediveo/bbeequeue"

	ctx, cancel := context.WithCancel(context.TODO())
	ch, err := bbeequeue.New[Foo](ctx, mymap)
	for {
	    event, ok := <-ch
	    if !ok {
	        break
	    }
	}

# Trivia

The package name “bbeequeue” is a terrible pun on eBPF's bee mascot, ringbuffers
or queues, and finally, burning things beyond recognition, also known as “BBQ”.
*/
package bbeequeue
