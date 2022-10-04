/*
This programm creates the compressed Release folder of
DDP from all locally built components

It consumes a config.json file which should look like this:

	{
		"Kompilierer": "<Directory to the Kompilierer repo>",
		"vscode-ddp": "<Directory to the vscode-ddp repo>",
		"DDPLS": "<Directory to the DDPLS repo>"
		"mingw": "<Directory to the mingw64 installation that should be shiped>"
	}

In the Kompilierer-repo, $make should have already been executed
*/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/DDP-Projekt/Installer/compression"
	cp "github.com/otiai10/copy"
	"github.com/spf13/viper"
)

func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

var compressExt = ".zip"

const outDir = "./DDP"

// read the config file
func init() {
	viper.SetConfigFile("config.json")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if runtime.GOOS == "linux" {
		compressExt = ".tar.gz"
	}
}

func main() {
	// cleanup from previous builds
	errPanic(os.RemoveAll(outDir))
	errPanic(os.RemoveAll(outDir + compressExt))

	// read the json file
	compDir := filepath.Join(viper.GetString("Kompilierer"), "build", "DDP")
	extDir := viper.GetString("vscode-ddp")
	lsDir := viper.GetString("DDPLS")
	mingwDir := viper.GetString("mingw")
	cwd, err := os.Getwd() // current working directory
	errPanic(err)

	// copy kddp
	errPanic(cp.Copy(compDir, outDir))
	// copy the extension output (.vsix file)
	runCmd(extDir, "vsce", "package", "-o", filepath.Join(cwd, outDir))
	// build the language server into the output directory
	runCmd(lsDir, "go", "build", "-o", filepath.Join(cwd, outDir, "bin"), ".")
	if runtime.GOOS == "windows" {
		// compress mingw and put it into the output directory
		errPanic(compression.CompressFolder(mingwDir, filepath.Join(outDir, "mingw64"+compressExt)))
	}
	// compress the output directory
	errPanic(compression.CompressFolder(outDir, outDir+compressExt))
}

func runCmd(dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("output of '$%s': %s\n", cmd, out)
		panic(err)
	}
}
