package main

import (
	"log"
	"os"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatalf("error: %v", err)
	}
}
