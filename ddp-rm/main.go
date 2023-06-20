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

// always exits with code 0 to not interrupt the makefile
func main() {
	for _, path := range os.Args[1:] {
		if err := os.Remove(path); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}
}
