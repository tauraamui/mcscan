package main

import (
	"errors"
	"fmt"
	stdos "os"
	"path/filepath"
	"strings"

	"github.com/hack-pad/hackpadfs/os"
	"github.com/tauraamui/mcscan/vfs"
)

// ~/Library/Application Support/minecraft

func main() {
	fsys := must(resolveFS(string(filepath.Separator)))

	configDirPath := must(stdos.UserConfigDir())
	configDirPath = strings.TrimPrefix(configDirPath, string(filepath.Separator))
	mcSavesPath := filepath.Join(configDirPath, "minecraft", "saves")
	fi := must(fsys.Stat(mcSavesPath))

	if fi.IsDir() {
		worldDirs := must(fsys.ReadDir(mcSavesPath))

		for _, wdir := range worldDirs {
			name := wdir.Name()
			if !vfs.IsHidden(name) {
				fmt.Printf(filepath.Join(configDirPath, mcSavesPath, name) + "\n")
			}
		}
	}
}

func resolveFS(base string) (*os.FS, error) {
	fs := os.NewFS()

	baseDirectory := must(fs.FromOSPath(base)) // Convert to an FS path
	baseDirFS := must(fs.Sub(baseDirectory))   // Run all file system operations rooted at the current working directory

	ofs, ok := baseDirFS.(*os.FS)
	if !ok {
		return nil, errors.New("sub FS not an OS instance FS")
	}

	return ofs, nil
}

func must[T any](v T, err error) T {
	if err != nil {
		_, _ = fmt.Fprintln(stdos.Stderr, err)
		stdos.Exit(1)
	}
	return v
}
