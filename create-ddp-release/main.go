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
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
	errPanic(os.Mkdir(filepath.Join(outDir, "vscode-ddp"), os.ModePerm))
	runCmd(extDir, "vsce", "package", "-o", filepath.Join(cwd, outDir, "vscode-ddp"))
	// build the language server into the output directory
	runCmd(lsDir, "go", "build", "-o", filepath.Join(cwd, outDir, "bin"), ".")
	if runtime.GOOS == "windows" {
		// compress mingw and put it into the output directory
		errPanic(os.Mkdir(filepath.Join(outDir, "gcc"), os.ModePerm))
		errPanic(compressFolder(mingwDir, filepath.Join(outDir, "gcc", "mingw64"+compressExt)))
	}
	// compress the output directory
	errPanic(compressFolder(outDir, outDir+compressExt))
}

func compressFolder(from, to string) error {
	// create the .zip/.tar file
	f, err := os.Create(to)
	if err != nil {
		return err
	}
	defer f.Close()

	if runtime.GOOS == "windows" {
		writer := zip.NewWriter(f)
		defer writer.Close()

		// go through all the files of the source
		return filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// create a local file header
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// set compression
			header.Method = zip.Deflate

			// set relative path of a file as the header name
			header.Name, err = filepath.Rel(filepath.Dir(from), path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				header.Name += "/"
			}

			// create writer for the file header and save content of the file
			headerWriter, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(headerWriter, f)
			return err
		})
	} else if runtime.GOOS == "linux" {
		// tar > gzip > file
		zr := gzip.NewWriter(f)
		tw := tar.NewWriter(zr)

		// walk through every file in the folder
		filepath.Walk(from, func(file string, fi os.FileInfo, _ error) error {
			// generate tar header
			header, err := tar.FileInfoHeader(fi, file)
			if err != nil {
				return err
			}

			header.Name, err = filepath.Rel(filepath.Dir(from), file)
			if err != nil {
				return err
			}

			// write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			// if not a dir, write file content
			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
			}
			return nil
		})

		// produce tar
		if err := tw.Close(); err != nil {
			return err
		}
		// produce gzip
		if err := zr.Close(); err != nil {
			return err
		}
		//
		return nil
	} else {
		panic("invalid OS")
	}
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
