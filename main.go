package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sokool/wpf/internal"
)

func main() {
	shutdown := context.TODO() // todo implement sig-calls to stop gracefully stop server

	if err := wpf.New(shutdown).Run(":8080"); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

}
