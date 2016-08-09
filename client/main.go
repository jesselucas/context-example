package main

import (
	"context"
	"fmt"
	"log"
	"net/rpc"
	"time"
)

func main() {
	serverAddr := "127.0.0.1"

	client, err := rpc.Dial("tcp", serverAddr+":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	// Create 10 workers on the server
	var reply bool
	err = client.Call("Work.Start", 10, &reply)
	if err != nil {
		log.Fatal("Work.Start error:", err)
	}

	if !reply {
		log.Fatal("Reply failed", reply)
	}

	// Block until the RPC call timeout happens
	select {
	case <-ctx.Done():
		// Stop all the workers we started
		err = client.Call("Work.Stop", struct{}{}, &reply)
		if err != nil {
			log.Fatal("stop error:", err)
		}

		client.Close()

		if err := ctx.Err(); err != nil {
			fmt.Println(err)
		}
	}
}
