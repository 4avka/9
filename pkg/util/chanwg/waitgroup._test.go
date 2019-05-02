package chanwg

import (
	"os"
	"testing"
	"time"
)

func TestWaitGroup(t *testing.T) {
	wg := New()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		println("added waitgroup worker")
		go func() {
			// println("hi, i'm a goroutine")
			// fmt.Println("current number of workers:", wg.WorkerCount())
			<-time.After(time.Second / 2)
			println("worker done")
			wg.Done(1)
		}()
	}
	go func() {
		<-time.After(time.Second)
		println("timing out")
		os.Exit(0)
	}()
	println("waiting for worker to finish")
	wg.Wait()
	println("done")
}
