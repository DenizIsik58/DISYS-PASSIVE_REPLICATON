package main

import (
	"ChatService/increment"
	"ChatService/util"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type IncrementServer struct {
	increment.UnimplementedIncrementServiceServer
	id int64
	lock    sync.Mutex
	counter int64
	clients  map[int]increment.IncrementServiceClient
	port int64
	leaderPort int64
}

var incrementServer *IncrementServer

func (s *IncrementServer) Increment(ctx context.Context, incrementer *increment.IncrementRequest) (*increment.IncrementResponse, error) {

	s.lock.Lock()
	defer s.lock.Unlock()


	var requestSucceeded bool
	var returnValue = s.counter

	if incrementer.GetIncrementValue() > 0 {
		log.Printf("You have successfully added the highest value %v", incrementer.GetIncrementValue())
		s.counter += incrementer.GetIncrementValue()
		requestSucceeded = true
	}else {
		log.Printf("The value is not greater than 0, and you may not give any negative numbers")
		requestSucceeded = false
	}

	return &increment.IncrementResponse{IncrementValue: returnValue, IsIncremented: requestSucceeded}, nil
}


func (s *IncrementServer) SetId(ctx context.Context, id *increment.SetIdResponse) (*increment.SetIdResponse, error) {
		s.id = id.GetId()
		log.Printf("Your ID is %v", id)
	return id, nil
}

func main() {

	port := strconv.Itoa(util.MinPort(10000))

	log.Print("Loading Incrementer service now...")

	listener, err := net.Listen("tcp", "localhost:" + port)

	if err != nil {
		log.Fatalf("TCP failed to listen... %s", err)
		return
	}

	log.Print("Listener registered - setting up server now...")

	incrementServer = &IncrementServer{
		id:                                  int64(rand.Intn(1000000000)),
		lock:                                sync.Mutex{},
		counter:                             0,
		clients:                             make(map[int]increment.IncrementServiceClient),
	}

	grpcServer := grpc.NewServer()

	increment.RegisterIncrementServiceServer(grpcServer, incrementServer)

	log.Print("===============================================================================")
	log.Print("                            Welcome to IncrementService                            ")
	log.Print("            Users can connect at any time and increment a value!            ")
	log.Print("===============================================================================")

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Failed to server gRPC server over port %v", port )
	}

	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	for {
		if incrementServer == nil {
			continue
		}

		id := incrementServer.id
		port := incrementServer.port

		for i := 1; i < 10; i++ {
			_port := i * 10000

			// We don't want our own server
			if _port == int(incrementServer.port) {
				continue
			}

			wg.Add(1)

			go func() {
				defer wg.Done()

				 client, ok := incrementServer.clients[int(port)]


				 if !ok {
					 log.Print("error has occurred")
				 }

				connectionAddress := fmt.Sprintf("localhost:%v", port)

				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()

				conn, err := grpc.DialContext(ctx, connectionAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

				if err != nil {
					log.Print("Couldn't host on port: ", port)
				}

				client = increment.NewIncrementServiceClient(conn)

				incrementServer.clients[int(port)] = client

				if client == nil {
					return
				}

				resp, err := client.SetId(context.Background(), &increment.SetIdResponse{Id: int64(rand.Intn(10000000000))})

				if err != nil {
					delete(incrementServer.clients, _port)
					return
				}

				lock.Lock()
				if resp.Id < id {
					port = int64(_port)
					id = resp.Id
				}
				lock.Unlock()
			}()
		}

		wg.Wait()

		// We were just made leader
		if incrementServer.port == port && incrementServer.leaderPort != port {
			log.Printf("%s >> I am now the primary server!", strconv.Itoa(int(port)))
			go func() {
				os.WriteFile("../port.txt", []byte(strconv.Itoa(int(port))), 0644)
			}()
		}

		incrementServer.leaderPort = port
	}
}
