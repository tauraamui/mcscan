package main

import (
	"errors"
	"fmt"
	stdos "os"
	"path/filepath"

	"github.com/hack-pad/hackpadfs/os"
	"github.com/tauraamui/mcscan/internal/scan"
)

func main() {
	fsys := must(resolveFS(string(filepath.Separator)))

	world := must(scan.OpenWorldByName(fsys, "DebugTestWorld"))
	fmt.Println(world.Name())

	world.BlocksCount()

	lvl := must(world.ReadLevel())
	lvl.Data.SpawnX = 125
	lvl.Data.SpawnY = 77
	lvl.Data.SpawnZ = 163

	lvl.Data.DayTime = 15000 // set time to night

	must(0, world.WriteLevel(lvl))
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
