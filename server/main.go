package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

// Work service handles canceling it's workers with a context
type Work struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

// Start workers and stop them if the context is canceled
func (w *Work) Start(workers int, ack *bool) error {
	fmt.Printf("Start %d workers \n", workers)
	for i := 0; i < workers; i++ {
		w.wg.Add(1)
		go func(worker int) {
			defer w.wg.Done()

			for {
				// some long task
				select {
				case <-w.ctx.Done():
					fmt.Println("Stopped worker", worker)
					return
				}
			}
		}(i)
	}
	*ack = true
	return nil
}

// Stop cancel all workers goroutines
func (w *Work) Stop(_ struct{}, ack *bool) error {
	fmt.Println("Stop all workers")
	w.cancel()
	if err := w.ctx.Err(); err != nil {
		*ack = false
		fmt.Println(err)
	}

	// decrement initial wg
	w.wg.Done()

	*ack = true
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Work{
		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
	}
	// initial wait to stop main from exiting
	w.wg.Add(1)

	rpc.Register(w)
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go func() { rpc.Accept(l) }()

	// Block until all workers stop
	w.wg.Wait()
	fmt.Println("All workers stopped. Shutting down server.")
}
