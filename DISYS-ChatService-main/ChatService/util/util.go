package util

import (
"context"
"math"
"strconv"
"sync"
"time"

"google.golang.org/grpc"
"google.golang.org/grpc/credentials/insecure"
)

func Min(l []int) int {
	min := math.MaxInt

	for _, value := range l {
		if value < min {
			min = value
		}
	}

	return min
}

func MinPort(multiplier int) int {
	ports := make([]int, 0)
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	for i := 1; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			defer wg.Done()

			_, err := grpc.DialContext(ctx, "localhost:"+strconv.Itoa(i*multiplier), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

			if err != nil {
				lock.Lock()
				ports = append(ports, i*multiplier)
				lock.Unlock()
			}
		}(i)
	}

	wg.Wait()

	return Min(ports)
}