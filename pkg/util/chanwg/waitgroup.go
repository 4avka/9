package chanwg

type WaitGroup struct {
	workers int
	ops     chan int
}

func New(bufsize int) *WaitGroup {
	wg := &WaitGroup{
		ops: make(chan int, bufsize),
	}
	go func() {
		// wait loop doesn't start until something is put into the ops chan
		done := false
		for !done {
			select {
			case op := <-wg.ops:
				wg.workers += op
				if wg.workers < 1 {
					done = true
					close(wg.ops)
				}
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
	wg.ops <- delta
}

// Done subtracts a non-negative value from the workers count
func (wg *WaitGroup) Done(delta int) {
	// println("worker finished")
	if delta < 0 {
		return
	}
	wg.ops <- -delta
}

// Wait blocks until the waitgroup decrements to zero
func (wg *WaitGroup) Wait() {
	for {
		op, ok := <-wg.ops
		if !ok {
			break
		} else {
			wg.ops <- op
		}
	}
}
