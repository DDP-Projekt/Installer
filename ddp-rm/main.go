/*
This program is a very simple implementation of the unix `rm` command.
It is shipped on windows for the installer to call `make clean` if the
runtime and stdlib had to be rebuilt
*/
package main

import (
	"fmt"
	"os"
)

func main() {
	exitStatus := 0
	for _, path := range os.Args[1:] {
		if err := os.Remove(path); err != nil {
			exitStatus = 1
			fmt.Fprintf(os.Stderr, "%s: %s", path, err)
		}
	}
	os.Exit(exitStatus)
}
