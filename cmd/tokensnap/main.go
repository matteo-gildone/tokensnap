package main

import (
	"fmt"
	"os"

	"github.com/matteo-gildone/tokensnap/internal/command"
)

var commands = []*command.Command{
	command.UpdateCmd,
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	for _, cmd := range commands {
		if cmd.Name == os.Args[1] {
			err := cmd.Run(os.Args[2:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%q failed: %v\n", cmd.Name, err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage tokensnap:\n")
	for _, cmd := range commands {
		fmt.Fprintf(os.Stderr, "  %s: %s\n", cmd.Name, cmd.Usage)
	}
}
