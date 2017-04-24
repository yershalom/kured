package main

import (
	"fmt"
	"log"
	"os"

	"github.com/weaveworks/kured"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("USAGE: %s <prometheusURL>", os.Args[0])
	}

	count, err := kured.CountActiveAlerts(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(count)
}
