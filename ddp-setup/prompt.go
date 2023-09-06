package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var scanner = bufio.NewScanner(os.Stdin)

func prompt(question string) bool {
	fmt.Print(ColorString(question+"? [ja/nein]: ", Cyan))
	scanner.Scan()
	answer := strings.ToLower(scanner.Text())
	return strings.ToLower(answer) == "ja"
}
