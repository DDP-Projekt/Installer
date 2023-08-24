/*
This programm creates the compressed Release folder of
DDP (DDP-<version-info>) from all locally built components.

It consumes a config.json (create-ddp-release/config.json) file which should look like this:

	{
		"Kompilierer": "<Directory to the Kompilierer repo>",
		"vscode-ddp": "<Directory to the vscode-ddp repo>",
		"DDPLS": "<Directory to the DDPLS repo>"
		"mingw": "<Directory to the mingw64 installation that should be shiped>"
	}

The "mingw" value only needs to be present on windows.
All the git-repos should be up-to-date.
*/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DDP-Projekt/Installer/compression"
	cp "github.com/otiai10/copy"
	"github.com/spf13/viper"
)

func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	compressExt = ".zip"
	ship_mingw  = true
)

// read the config file
func setup_config() {
	fmt.Println("reading config")
	viper.SetConfigFile("config.json")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if runtime.GOOS == "linux" {
		compressExt = ".tar.gz"
	}
}

func gen_out_dir(ddpBuildDir string) string {
	version_path := "DDP-" + strings.ReplaceAll(
		strings.TrimPrefix(
			strings.TrimRight(
				strings.Split(
					runCmd(filepath.Join(ddpBuildDir, "bin"), "kddp", "version"), "\n",
				)[0],
				"\r\n",
			),
			"DDP Version: ",
		),
		" ",
		"-",
	)
	if runtime.GOOS == "windows" && viper.GetString("mingw") == "" {
		version_path += "-no-mingw"
		ship_mingw = false
	}

	// delete version path if it exists
	if _, err := os.Stat(version_path); !os.IsNotExist(err) {
		fmt.Printf("deleting %s\n", version_path)
		errPanic(os.RemoveAll(version_path))
	}

	return version_path
}

func main() {
	setup_config()

	// read the json file
	compDir := filepath.Join(viper.GetString("Kompilierer"))
	ddpBuildDir := filepath.Join(compDir, "build", "DDP")
	extDir := viper.GetString("vscode-ddp")
	lsDir := viper.GetString("DDPLS")
	mingwDir := viper.GetString("mingw")
	cwd, err := os.Getwd() // current working directory
	errPanic(err)

	// compile and copy kddp
	runCmd(compDir, "make", "all")
	outDir := gen_out_dir(ddpBuildDir)
	fmt.Println("copying Kompilierer/build/DDP directory")
	errPanic(cp.Copy(ddpBuildDir, outDir))
	// copy the extension output (.vsix file)
	runCmd(extDir, "vsce", "package", "-o", filepath.Join(cwd, outDir, "vscode-ddp.vsix"))
	// build the language server into the output directory
	runCmd(lsDir, "go", "build", "-o", filepath.Join(cwd, outDir, "bin"), ".")
	if runtime.GOOS == "windows" {
		// build ddp-rm
		runCmd("../ddp-rm/", "go", "build", "-o", filepath.Join(cwd, outDir, "bin"), ".")
		if ship_mingw {
			// compress mingw and put it into the output directory
			fmt.Println("compressing mingw")
			errPanic(compression.CompressFolder(filepath.Clean(mingwDir), filepath.Join(outDir, "mingw64"+compressExt)))
		}
	}
	// build the installer
	runCmd("../ddp-setup/", "go", "build", "-o", filepath.Join(cwd, outDir), ".")
	// compress the output directory
	fmt.Println("compressing release folder")
	errPanic(compression.CompressFolder(filepath.Clean(outDir), filepath.Clean(outDir+compressExt)))
}

func runCmd(dir, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	fmt.Printf("Running cmd %s in dir %s\n", cmd.String(), cmd.Dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("output of '$%s': %s\n", cmd, out)
		panic(err)
	}
	return string(out)
}
