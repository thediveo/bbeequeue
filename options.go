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

import "errors"

// Option configures optional settings.
type Option func(o *options) error

type options struct {
	size  int
	errch chan error
}

// WithSize configures the buffer size of the channel returned by [New]. 0
// configures an unbuffered channel. New will return an error when trying to
// configure a size less than zero.
func WithSize(size int) Option {
	return func(o *options) error {
		if size < 0 {
			return errors.New("channel size must be greater or equal to zero")
		}
		o.size = size
		return nil
	}
}

// WithErrorChannel configures a caller-supplied error channel on which
// ringbuffer receive und data unmarshalling errors are reported. This channel
// will not be automatically closed.
func WithErrorChannel(errch chan error) Option {
	return func(o *options) error {
		o.errch = errch
		return nil
	}
}
