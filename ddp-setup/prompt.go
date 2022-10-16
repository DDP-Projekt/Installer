package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var scanner = bufio.NewScanner(os.Stdin)

func prompt(question string) bool {
	fmt.Print(ColorString(question+"? [y/n]: ", Cyan))
	scanner.Scan()
	answer := scanner.Text()
	return strings.ToLower(answer) == "y"
}
