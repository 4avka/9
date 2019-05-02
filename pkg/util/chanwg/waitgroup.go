package chanwg

import "fmt"

type WaitGroup struct {
	workers uint
	ops     chan func()
	ready   chan struct{}
	done    chan struct{}
}

func New() *WaitGroup {
	wg := &WaitGroup{
		ops:   make(chan func()),
		done:  make(chan struct{}),
		ready: make(chan struct{}),
	}
	go func() {
		// wait loop doesn't start until something is put into thte
		done := false
		for !done {
			select {
			case fn := <-wg.ops:
				println("received op")
				fn()
				fmt.Println("num workers:", wg.WorkerCount())
				// if !(wg.workers < 1) {
				// 	println("wait counter at zero")
				// 	done = true
				// 	close(wg.done)
				// }
			default:
			}
		}

	}()
	return wg
}

// Add adds a non-negative number
func (wg *WaitGroup) Add(delta int) {
	if delta < 0 {
		return
	}
	fmt.Println("adding", delta, "workers")
	wg.ops <- func() {
		wg.workers += uint(delta)
	}
}

// Done subtracts a non-negative value from the workers count
func (wg *WaitGroup) Done(delta int) {
	println("worker finished")
	if delta < 0 {
		return
	}
	println("pushing op to channel")
	wg.ops <- func() {
		println("finishing")
		wg.workers -= uint(delta)
	}
	// println("op should have cleared by now")
}

// Wait blocks until the waitgroup decrements to zero
func (wg *WaitGroup) Wait() {
	println("a worker is waiting")
	<-wg.done
	println("job done")
}

func (wg *WaitGroup) WorkerCount() int {
	return int(wg.workers)
}
