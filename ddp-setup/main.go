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
	gccCmd    = "gcc"
	makeCmd   = "make"
	arCmd     = "ar"
	vscodeCmd = "code"
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
		gccCmd = filepath.ToSlash(gccCmd)
		arCmd, err = filepath.Abs(filepath.Join("mingw64", "bin", "ar"))
		if err != nil {
			WarnF("error getting absolute Path: %s", err)
		}
		arCmd = filepath.ToSlash(arCmd)
		makeCmd, err = filepath.Abs(filepath.Join("mingw64", "bin", "mingw32-make"))
		if err != nil {
			WarnF("error getting absolute Path: %s", err)
		}
		makeCmd = filepath.ToSlash(makeCmd)

		DoneF("using unzipped mingw64 for gcc, ar and make")
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
			makeCmd = filepath.ToSlash(makeCmd)
		}
	}

	if isSameGccVersion() {
		DoneF("gcc versions match")
	} else {
		InfoF("re-building runtime and stdlib")
		recompileLibs()
	}

	hasVscode := false
	if vscodeCmd, hasVscode = LookupCommand(vscodeCmd); hasVscode {
		InfoF("installing vscode-ddp as vscode extension")
		vsixPath := findVSIXFile()
		if vsixPath == "" {
			WarnF("No .vsix file found, aborting installation of vscode-ddp")
		} else {
			runCmd("", vscodeCmd, "--install-extension", vsixPath)
		}
	}

	InfoF("Press ENTER to exit...")
	fmt.Scanln()
}

func isSameGccVersion() bool {
	gccVersion, err := runCmd("", gccCmd, "-dumpfullversion")
	if err != nil {
		return false
	}
	gccVersion = strings.Trim(gccVersion, "\n") // TODO: this
	kddpVersionOutput, err := runCmd("", filepath.Join("bin", "kddp"), "version")
	if err != nil {
		return false
	}
	gccVersionLine := strings.Split(kddpVersionOutput, "\n")[2]
	kddpGccVersion := strings.Trim(strings.Split(gccVersionLine, " ")[2], "\n")
	match := gccVersion == kddpGccVersion
	if !match {
		InfoF("local gcc version, and kddp gcc version mismatch (%s vs %s)", gccVersion, kddpGccVersion)
	}
	return match
}

func recompileLibs() {
	if runtime.GOOS == "linux" {
		if _, err := runCmd("lib/runtime/", makeCmd); err != nil {
			return
		}
		DoneF("re-compiled the runtime")
		if _, err := runCmd("lib/stdlib/", makeCmd); err != nil {
			return
		}
		DoneF("re-compiled the stdlib")
	} else if runtime.GOOS == "windows" {
		compileLibsWindows()
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
	if err := cp.Copy("lib/runtime/libddpruntime.a", "lib/libddpruntime.a"); err != nil {
		ErrorF("error copying re-compiled runtime: %s", err)
	}
	InfoF("copying re-compiled stdlib")
	if err := cp.Copy("lib/stdlib/libddpstdlib.a", "lib/libddpstdlib.a"); err != nil {
		ErrorF("error copying re-compiled stdlib: %s", err)
	}

	InfoF("cleaning runtime directory")
	if err := os.Remove("lib/runtime/libddpruntime.a"); err != nil {
		WarnF("error while cleaning runtime directory: %s", err)
	}
	if err := removeObjects("lib/runtime/"); err != nil {
		WarnF("error while cleaning runtime directory: %s", err)
	}
	InfoF("cleaning stdlib directory")
	if err := os.Remove("lib/stdlib/libddpstdlib.a"); err != nil {
		WarnF("error while cleaning stdlib directory: %s", err)
	}
	if err := removeObjects("lib/stdlib/"); err != nil {
		WarnF("error while cleaning stdlib directory: %s", err)
	}

	DoneF("recompiled libraries")
}

func compileLibsWindows() {
	getFiles := func(dir, ext string) []string {
		files := make([]string, 0)

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() || err != nil {
				return nil
			}
			if filepath.Ext(path) == ext {
				files = append(files, strings.TrimPrefix(filepath.ToSlash(path), dir))
			}
			return nil
		})

		if err != nil {
			WarnF("error while filtering files: %s", err)
		}

		return files
	}

	sources := getFiles("lib/runtime/", ".c")
	objects := Map(sources, func(str string) string {
		return filepath.Base(strings.ReplaceAll(str, ".c", ".o"))
	})
	args := append(make([]string, 0), "-c", "-Wall", "-Wno-format", "-O2", "-I./include/")
	args = append(args, sources...)
	if _, err := runCmd("lib/runtime", gccCmd, args...); err != nil {
		return
	}
	args = append(make([]string, 0), "cr", "libddpruntime.a")
	args = append(args, objects...)
	if _, err := runCmd("lib/runtime", arCmd, args...); err != nil {
		return
	}
	DoneF("re-compiled the runtime")

	sources = getFiles("lib/stdlib/", ".c")
	objects = Map(sources, func(str string) string {
		return filepath.Base(strings.ReplaceAll(str, ".c", ".o"))
	})
	args = append(make([]string, 0), "-c", "-Wall", "-Wno-format", "-O2", "-I./include/", "-I../runtime/include/")
	args = append(args, sources...)
	if _, err := runCmd("lib/stdlib", gccCmd, args...); err != nil {
		return
	}
	args = append(make([]string, 0), "cr", "libddpstdlib.a")
	args = append(args, objects...)
	if _, err := runCmd("lib/stdlib", arCmd, args...); err != nil {
		return
	}
	DoneF("re-compiled the stdlib")
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
		DoneF("Found %s in %s", cmd, path)
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

func Map[T any](s []T, mapFunc func(t T) T) []T {
	result := make([]T, 0, len(s))
	for _, v := range s {
		result = append(result, mapFunc(v))
	}
	return result
}

func findVSIXFile() (result string) {
	if err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ".vsix" {
			result = path
		}
		return nil
	}); err != nil {
		ErrorF("Error searching for .vsix file: %s", err)
	}
	return result
}
