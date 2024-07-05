package main

import (
	"fmt"
	. "github.com/jsp-lqk/metapipe-memcached/internal"
	"strconv"
	"sync"
)

func main() {
	client, err := NewBaseClient("127.0.0.1", 11211, 100)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	var wg sync.WaitGroup

	// Loop 50 times
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ok, err := client.Set(strconv.Itoa(i), []byte(fmt.Sprintf("value-%d", i)), 0)
			if err != nil {
				fmt.Println("Error:", err.Error())
			}
			if !ok {
				fmt.Println("store failed")
			} else {
				fmt.Println("set ok")
			}
		}(i)
	}

	if _, err = client.Stale("a"); err != nil {
		fmt.Println("Error:", err.Error())
	}

	for i := 50; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ok, err := client.Set(strconv.Itoa(i), []byte(fmt.Sprintf("value-%d", i)), 0)
			if err != nil {
				fmt.Println("Error:", err.Error())
			}
			if !ok {
				fmt.Println("store failed")
			} else {
				fmt.Println("set ok")
			}
		}(i)
	}

	wg.Wait()

	var wt sync.WaitGroup

	for i := 0; i < 10; i++ {
		wt.Add(1)
		go func(i int) {
			defer wt.Done()
			r, err := client.Get(strconv.Itoa(i))
			if err != nil {
				fmt.Println("Error:", err.Error())
			}
			fmt.Printf("Iteration %d received value: %s\r\n", i, string(r))
		}(i)
	}
/*
	if _, err = client.Stale("a"); err != nil {
		fmt.Println("Error:", err.Error())
	}*/
	
	
	wt.Wait()
}
