package main

import (
	"ChatService/increment"
	"bufio"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

//	response, _ := client.SendMessage(context.Background(), &increment.Message{Content: "Hello from the client!"})
//	log.Printf("Response from server: %s", response.Content)


var client increment.IncrementServiceClient

func Increment(value int64) {

		resp, err := client.Increment(context.TODO(), &increment.IncrementRequest{IncrementValue: value})

		if resp.GetIsIncremented() == false {
			log.Print("Your request for incrementing a value wasn't high enough. The highest value is:")
		}
		if err != nil {
			log.Fatalf("An error occured during the process of increment the value")
		}

		log.Print("You have increment the current counter. The value before was", resp.GetIncrementValue())
	}

func FindFrontEnd() {
	lock := sync.Mutex{}
	found := false
	foundCh := make(chan bool)

	for i := 0; i < 10; i++ {
		port := i * 1000
		go func() {
			address := fmt.Sprintf("localhost:%v", port)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

			if err != nil || found {
				return
			}

			lock.Lock()
			client = increment.NewIncrementServiceClient(conn)
			found = true
			foundCh <- true
			lock.Unlock()
		}()
	}

	select {
	case <-foundCh:
	case <-time.After(2 * time.Second):
		return
	}
}


func main() {

	FindFrontEnd()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		val, _ := strconv.ParseInt(scanner.Text(), 10,64)
		Increment(val)
	}
}
