package main

import (
	"errors"
	"fmt"
	stdos "os"
	"path/filepath"
	"strings"

	"github.com/hack-pad/hackpadfs/os"
	"github.com/tauraamui/mcscan/internal/vfs"
	mc "github.com/tauraamui/mcscan/pkg/minecraft"
)

func main() {
	// Acquire file system access which starts from root
	// rather than one which starts from user config dir.
	fsys := must(resolveFS(string(filepath.Separator)))

	configDirPath := must(stdos.UserConfigDir())
	configDirPath = strings.TrimPrefix(configDirPath, string(filepath.Separator))
	mcSavesPath := filepath.Join(configDirPath, "minecraft", "saves")
	fi := must(fsys.Stat(mcSavesPath))

	if fi.IsDir() {
		worldDirs := must(fsys.ReadDir(mcSavesPath))

		for _, wdir := range worldDirs {
			if !wdir.IsDir() {
				continue
			}
			name := wdir.Name()
			if !vfs.IsHidden(name) {
				wdirFullPath := filepath.Join(mcSavesPath, name)
				world := must(mc.OpenWorld(fsys, wdirFullPath))

				fmt.Println(world.Name())
				fmt.Println(world.RegionsCount())

				must(0, world.Close())
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
