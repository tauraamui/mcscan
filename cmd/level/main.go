package main

import (
	"errors"
	"fmt"
	stdos "os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/hack-pad/hackpadfs/os"
	mc "github.com/tauraamui/mcscan/pkg/minecraft"
)

type args struct {
	WorldName string `arg:"required"`
}

func (args) Version() string {
	return "mcscan v0.0.0"
}

func main() {
	var args args
	arg.MustParse(&args)

	fsys := must(resolveFS(string(filepath.Separator)))

	world, err := mc.OpenWorldByName(fsys, args.WorldName)
	if errors.Is(err, stdos.ErrNotExist) {
		fmt.Fprintf(stdos.Stderr, "could not find world data for '%s'\n", args.WorldName)
		stdos.Exit(1)
	}

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
