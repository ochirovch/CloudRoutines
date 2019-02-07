package main

import (
	"fmt"
	"time"

	"github.com/ochirovch/CollyRoutines/client"
)

func main() {
	for {

		bucket, err := client.GetListURLs("127.0.0.1/channel/get")
		if err != nil {
			time.Sleep(30 * time.Second)
		}
		if len(bucket.Paths) == 0 {
			time.Sleep(30 * time.Second)
		}
		for _, element := range bucket.Paths {
			fmt.Println(element)
		}
		client.SendResults()
	}
}
