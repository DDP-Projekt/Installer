package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var (
	scanner    = bufio.NewScanner(os.Stdin)
	always_yes = false
)

func prompt(question string) bool {
	if always_yes {
		return true
	}

	fmt.Print(ColorString(question+"? [y/n]: ", Cyan))
	scanner.Scan()
	answer := strings.ToLower(scanner.Text())
	return strings.ToLower(answer) == "y"
}
