package chanwg

type WaitGroup struct {
	started bool
	workers int
	ops     chan int
}

// Add adds a non-negative number
func (wg *WaitGroup) Add(delta int) {
	if delta < 0 {
		return
	}
	if wg.started {
		wg.ops <- delta
	} else {
		wg.ops = make(chan int)
		go func() {
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
		wg.ops <- delta
		wg.started = true
	}
}

// Done subtracts a non-negative value from the workers count
func (wg *WaitGroup) Done(delta int) {
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
