package main

import (
	"fmt"

	"github.com/DDP-Projekt/Installer/compression"
)

func main() {
	if err := compression.DecompressFolder("./create-ddp-release/DDP/mingw64.zip", "mingw64"); err != nil {
		fmt.Println(err)
	}
}
