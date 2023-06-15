package main

import (
	"errors"
	"fmt"
	stdos "os"
	"path/filepath"

	"github.com/hack-pad/hackpadfs/os"
	"github.com/tauraamui/mcscan/internal/scan"
	"github.com/tauraamui/mcscan/vfs"
)

// ~/Library/Application Support/minecraft

func main() {
	configDirPath := must(stdos.UserConfigDir())
	fsys := must(resolveFS(configDirPath))

	mcSavesPath := filepath.Join("minecraft", "saves")
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

func resolveWorlds(fsResolver func() (*os.FS, error)) []scan.World {
	return nil
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
