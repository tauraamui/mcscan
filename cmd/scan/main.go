package main

import (
	"errors"
	"fmt"
	stdos "os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/hack-pad/hackpadfs/os"
	mc "github.com/tauraamui/mcscan/pkg/minecraft"
)

type args struct {
	WorldPath string `arg:"--path"`
	WorldName string `arg:"--name"`
}

func (args) Version() string {
	return "mcutils v0.0.0"
}

func main() {
	var args args
	p := arg.MustParse(&args)
	execscan(args, p)
}

func execscan(args args, p *arg.Parser) {
	args.WorldPath = strings.Trim(args.WorldPath, string(filepath.Separator))

	if len(args.WorldName) == 0 && len(args.WorldPath) == 0 {
		p.Fail("must provide either --path or --name")
	}

	var nameDefined, pathDefined bool
	if len(args.WorldName) > 0 {
		nameDefined = true
	}

	if len(args.WorldPath) > 0 {
		pathDefined = true
	}

	if nameDefined && pathDefined {
		p.Fail("provide either both --path or --name not both")
	}

	if p.Subcommand() == nil {
		p.Fail("missing subcommand")
	}

	if nameDefined {
		if err := runCmd(args.WorldName, mc.OpenWorldByName, p.Subcommand()); err != nil {
			exit(err.Error())
		}
		return
	}

	if err := runCmd(args.WorldPath, mc.OpenWorld, p.Subcommand()); err != nil {
		exit(err.Error())
	}
}

func resolveFS(base string) (*os.FS, error) {
	fs := os.NewFS()

	baseDirectory, err := fs.FromOSPath(base) // Convert to an FS path
	if err != nil {
		return nil, err
	}

	baseDirFS, err := fs.Sub(baseDirectory) // Run all file system operations rooted at the current working directory
	if err != nil {
		return nil, err
	}

	ofs, ok := baseDirFS.(*os.FS)
	if !ok {
		return nil, errors.New("sub FS not an OS instance FS")
	}

	return ofs, nil
}

func runCmd(worldRef string, worldResolver mc.WorldResolver, subCmd any) error {
	fsys, err := resolveFS(string(filepath.Separator))
	if err != nil {
		return err
	}

	world, err := worldResolver(fsys, worldRef)

	if err != nil {
		if errors.Is(err, stdos.ErrNotExist) {
			return fmt.Errorf("could not find world data for '%s'", worldRef)
		}
		return err
	}

	return nil
}

func exit(format string, a ...any) {
	fmt.Fprintf(stdos.Stderr, format+"\n", a...)
	stdos.Exit(1)
}
