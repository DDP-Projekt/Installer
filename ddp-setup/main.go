package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/DDP-Projekt/Installer/compression"
)

var (
	gccCmd  = "gcc"
	makeCmd = "make"

	nativeGcc = true // wether we use a system-installed gcc or the zipped mingw64
)

func main() {
	_, hasGcc := LookupCommand("gcc")

	if !hasGcc && runtime.GOOS == "windows" {
		InfoF("gcc not found, unzipping mingw64")
		err := compression.DecompressFolder("mingw64.zip", "mingw64")
		if err != nil {
			ErrorF("Error while unzipping mingw64: %s", err)
			ErrorF("no gcc available, aborting")
			os.Exit(1)
		}

		gccCmd, err = filepath.Abs(filepath.Join("mingw64", "bin", "gcc.exe"))
		if err != nil {
			WarnF("error getting absolute Path: %s", err)
		}
		nativeGcc = false

		makeCmd, err = filepath.Abs(filepath.Join("mingw64", "bin", "mingw32-make.exe"))
		if err != nil {
			WarnF("error getting absolute Path: %s", err)
		}
		InfoF("using unzipped mingw64 for gcc and make")
	} else if !hasGcc && runtime.GOOS != "windows" {
		ErrorF("gcc not found, aborting")
		os.Exit(1)
	}

	if makeCmd == "make" { // if we don't use the zipped mingw32-make
		_, hasMake := LookupCommand("make")

		if !hasMake && runtime.GOOS == "windows" {
			InfoF("make not found, looking for mingw32-make")
			makeCmd, hasMake = LookupCommand("mingw32-make")
			if !hasMake {
				ErrorF("mingw32-make not found, aborting")
				os.Exit(1)
			}
		}
	}

	InfoF("Press ENTER to exit...")
	fmt.Scanln()
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmdStr := cmd.String()
	out, err := cmd.CombinedOutput()
	if err != nil {
		ErrorF("'%s' failed (%s) output: %s", cmdStr, err, out)
	}
	return err
}

func LookupCommand(cmd string) (string, bool) {
	InfoF("Looking for %s", cmd)
	path, err := exec.LookPath(cmd)
	if err == nil {
		InfoF("Found %s in %s", cmd, path)
	} else {
		WarnF("Unable to find %s", cmd)
	}
	return path, err == nil
}
