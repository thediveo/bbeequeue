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
	. "github.com/onsi/gomega/gleak"
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

		goodgoos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgoos))
		})
	})

	Context("unhappy times", func() {

		It("rejects a negative size", func(ctx context.Context) {
			Expect(bbeequeue.New[bbqEvent](ctx, objs.Events, bbeequeue.WithSize(-1))).
				Error().To(HaveOccurred())
		})

		It("rejects an incorrect map type", func(ctx context.Context) {
			Expect(bbeequeue.New[bbqEvent](ctx, objs.MapMen)).Error().To(HaveOccurred())
		})

		It("closes the event channel when its context gets done", func(ctx context.Context) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			ch := Successful(
				bbeequeue.New[bbqEvent](ctx, objs.Events, bbeequeue.WithSize(5)))

			cancel()
			Eventually(ch).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
				Should(BeClosed())
		})

		It("stops trying to send on the blocked channel when the context gets cancelled", func(ctx context.Context) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			By("creating an *unbuffered* event channel")
			_ = Successful(bbeequeue.New[bbqEvent](ctx, objs.Events))

			By("stuffing something into the ringbuffer in kernel space")
			Expect(objs.Emit.Run(&ebpf.RunOptions{
				Context: uint64(42),
			})).To(BeZero())

			By("waiting for the user space event pump time to get stuck")
			// We here make create use of Gomega's go routine leak checking:
			// this time, here we don't check for leaks but instead sample the
			// go routines until we see that the ringbuffer record pump is stuck
			// on either finally getting rid of its event successfully or the
			// context gets cancelled ... that is, the record pump is sitting in
			// a select and not moving anymore.
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(20 * time.Millisecond).
				Should(ContainElement(And(
					HaveField("State", "select"),
					HaveField("TopFunction", MatchRegexp(`/bbeequeue\.New\[.*\]\.func2$`)))))
			By("cancelling the context while the event pump is stuck")
			cancel()
			By("expecting the goroutines not to leak due to a stuck event pump")
		})

	})

	It("receives multiple events", func(ctx context.Context) {
		ch := Successful(
			bbeequeue.New[bbqEvent](ctx, objs.Events, bbeequeue.WithSize(5)))

		values := []uint64{42, 123, 666}

		for _, val := range values {
			Expect(objs.Emit.Run(&ebpf.RunOptions{
				Context: val,
			})).To(BeZero())
		}

		var ev bbqEvent
		for _, val := range values {
			Eventually(ch).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
				Should(Receive(&ev))
			Expect(ev.Magic).To(Equal(val))
			Expect(ev.InverseMagic).To(Equal(^val))
		}
	})

	It("reports unmarshalling problems", func(ctx context.Context) {
		errch := make(chan error)
		_ = Successful(
			bbeequeue.New[[32]byte](ctx, objs.Events, bbeequeue.WithErrorChannel(errch)))

		Expect(objs.Emit.Run(&ebpf.RunOptions{
			Context: uint64(666),
		})).To(BeZero())

		var err error
		Eventually(errch).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(Receive(&err))
		Expect(err).To(MatchError("unexpected EOF"))
	})

})
