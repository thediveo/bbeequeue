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

package bbeequeue_test

import (
	"context"
	"os"
	"time"

	"github.com/cilium/ebpf"
	"github.com/thediveo/bbeequeue"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("bbeequeues", func() {

	var objs bbqObjects

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		Expect(loadBbqObjects(&objs, nil)).To(Succeed())
		DeferCleanup(func() {
			Expect(objs.Close()).To(Succeed())
		})
	})

	It("rejects an invalid map", func(ctx context.Context) {
		Expect(bbeequeue.New[bbqEvent](ctx, objs.MapMen, 0)).Error().To(HaveOccurred())
	})

	It("closes the event channel when its context gets done", func(ctx context.Context) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		ch := Successful(bbeequeue.New[bbqEvent](ctx, objs.Events, 5))

		cancel()
		Eventually(ch).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(BeClosed())
	})

	It("receives events", func(ctx context.Context) {
		ch := Successful(bbeequeue.New[bbqEvent](ctx, objs.Events, 5))

		Expect(objs.Emit.Run(&ebpf.RunOptions{
			Context: uint64(42),
		})).To(BeZero())

		Expect(objs.Emit.Run(&ebpf.RunOptions{
			Context: uint64(666),
		})).To(BeZero())

		var ev bbqEvent
		for _, val := range []uint64{42, 666} {
			Eventually(ch).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
				Should(Receive(&ev))
			Expect(ev.Magic).To(Equal(val))
			Expect(ev.InverseMagic).To(Equal(^val))
		}
	})

})
