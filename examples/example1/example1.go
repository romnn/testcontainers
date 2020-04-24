package main

import (
	"fmt"

	"github.com/romnnn/testcontainers"
)

func run() string {
	return testcontainers.Shout("This is an example")
}

func main() {
	fmt.Println(run())
}
