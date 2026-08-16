package main

import (
	"fmt"
	"os"

	"github.com/gspaim/Runtgine/internal/entrypoint/cli"
)

func main() {
	if err := cli.NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
