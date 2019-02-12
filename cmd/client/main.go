package main

import (
	"fmt"
	"time"

	"github.com/ochirovch/CloudRoutines/client"
)

func main() {
	for {

		bucket, err := client.GetTask()
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
