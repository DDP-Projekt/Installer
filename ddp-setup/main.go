package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DDP-Projekt/Installer/compression"
	cp "github.com/otiai10/copy"
)

var (
	gccCmd  = "gcc"
	makeCmd = "make"
	arCmd   = "ar"
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

		gccCmd, err = filepath.Abs(filepath.Join("mingw64", "bin", "gcc"))
		if err != nil {
			WarnF("error getting absolute Path: %s", err)
		}
		arCmd, err = filepath.Abs(filepath.Join("mingw64", "bin", "ar"))
		if err != nil {
			WarnF("error getting absolute Path: %s", err)
		}
		makeCmd, err = filepath.Abs(filepath.Join("mingw64", "bin", "mingw32-make"))
		if err != nil {
			WarnF("error getting absolute Path: %s", err)
		}

		InfoF("using unzipped mingw64 for gcc, ar and make")
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

	if isSameGccVersion() {
		InfoF("gcc versions match")
	} else {
		InfoF("local gcc version, and kddp gcc version mismatch")
		InfoF("re-building runtime and stdlib")
		recompileLibs()
	}

	InfoF("Press ENTER to exit...")
	fmt.Scanln()
}

func isSameGccVersion() bool {
	gccVersion, err := runCmd("", gccCmd, "-dumpfullversion")
	if err != nil {
		return false
	}
	kddpVersionOutput, err := runCmd("", filepath.Join("bin", "kddp"), "version")
	if err != nil {
		return false
	}
	gccVersionLine := strings.Split(kddpVersionOutput, "\n")[2]
	kddpGccVersion := strings.Split(gccVersionLine, " ")[2]
	return gccVersion == kddpGccVersion
}

func recompileLibs() {
	if _, err := runCmd("lib/runtime", makeCmd, "build", fmt.Sprintf("CC=%s", gccCmd), fmt.Sprintf("AR=%s", arCmd)); err != nil {
		return
	}
	if _, err := runCmd("lib/stdlib", makeCmd, "build", fmt.Sprintf("CC=%s", gccCmd), fmt.Sprintf("AR=%s", arCmd)); err != nil {
		return
	}

	InfoF("removing pre-compiled runtime")
	if err := os.Remove("lib/libddpruntime.a"); err != nil {
		WarnF("error removing pre-compiled runtime: %s", err)
	}
	InfoF("removing pre-compiled stdlib")
	if err := os.Remove("lib/libddpstdlib.a"); err != nil {
		WarnF("error removing pre-compiled stdlib: %s", err)
	}

	InfoF("copying re-compiled runtime")
	if err := cp.Copy("lib/runtime/libddpruntime.a", "lib/"); err != nil {
		ErrorF("error copying re-compiled runtime: %s", err)
	}
	InfoF("copying re-compiled stdlib")
	if err := cp.Copy("lib/stdlib/libddpstdlib.a", "lib/"); err != nil {
		ErrorF("error copying re-compiled stdlib: %s", err)
	}

	InfoF("cleaning runtime directory")
	if err := removeObjects("lib/runtime/"); err != nil {
		WarnF("error while cleaning runtime directory: %s", err)
	}
	InfoF("cleaning stdlib directory")
	if err := removeObjects("lib/stdlib/"); err != nil {
		WarnF("error while cleaning stdlib directory: %s", err)
	}
}

func runCmd(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmdStr := cmd.String()
	out, err := cmd.CombinedOutput()
	if err != nil {
		ErrorF("'%s' failed (%s) output: %s", cmdStr, err, out)
	}
	return string(out), err
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

func removeObjects(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || err != nil {
			return nil
		}

		if filepath.Ext(path) == ".o" {
			if err := os.Remove(path); err != nil {
				WarnF("Error removing '%s': %s", path, err)
			}
		}

		return nil
	})
}
