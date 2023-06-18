package main

import (
	"encoding/json"
	"errors"
	"fmt"
	stdos "os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/hack-pad/hackpadfs/os"
	mc "github.com/tauraamui/mcscan/pkg/minecraft"
)

type args struct {
	WorldPath string   `arg:"--path"`
	WorldName string   `arg:"--name"`
	ViewCmd   *ViewCmd `arg:"subcommand:view" help:"display world level data as JSON"`
	EditCmd   *EditCmd `arg:"subcommand:edit" help:"update level data fields to given values"`
}

func (args) Version() string {
	return "mcscan v0.0.0"
}

func main() {
	var args args
	p := arg.MustParse(&args)
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

	switch cmd := subCmd.(type) {
	case *ViewCmd:
		return viewCmd(world)
	case *EditCmd:
		return editCmd(world, cmd)
	}

	/*
		lvl := must(world.ReadLevel())
		lvl.Data.SpawnX = 125
		lvl.Data.SpawnY = 77
		lvl.Data.SpawnZ = 163

		lvl.Data.DayTime = 15000 // set time to night

		must(0, world.WriteLevel(lvl))
	*/

	return nil
}

type EditCmd struct {
	Values map[string]string
}

func editCmd(world *mc.World, cmd *EditCmd) error {
	lvl, err := world.ReadLevel()
	if err != nil {
		return err
	}

	lvlJSON, err := json.Marshal(lvl)
	if err != nil {
		return err
	}

	var lvlData map[string]any
	if err := json.Unmarshal(lvlJSON, &lvlData); err != nil {
		return err
	}

	for editValKey, editValVal := range cmd.Values {
		existingVal, ok := lvlData[editValKey]
		if !ok {
			return fmt.Errorf("property %s not found", editValKey)
		}

		switch existingVal.(type) {
		case string:
			lvlData[editValKey] = editValVal
		case int:
			valAsInt, err := strconv.Atoi(editValVal)
			if err != nil {
				return err
			}
			lvlData[editValKey] = valAsInt
		case float32:
			valAsFloat, err := strconv.ParseFloat(editValVal, 32)
			if err != nil {
				return err
			}
			lvlData[editValKey] = valAsFloat
		case float64:
			valAsFloat, err := strconv.ParseFloat(editValVal, 64)
			if err != nil {
				return err
			}
			lvlData[editValKey] = valAsFloat
		case bool:
			valAsBool, err := strconv.ParseBool(editValVal)
			if err != nil {
				return err
			}
			lvlData[editValKey] = valAsBool
		}
	}

	lvlDataBytes, err := json.Marshal(lvlData)
	if err != nil {
		return err
	}

	editedLvl := mc.Level{}
	if err := json.Unmarshal(lvlDataBytes, &editedLvl); err != nil {
		return err
	}

	if err := world.WriteLevel(&editedLvl); err != nil {
		return err
	}

	return nil
}

type ViewCmd struct{}

func viewCmd(world *mc.World) error {
	lvl, err := world.ReadLevel()
	if err != nil {
		return err
	}

	lvlJSON, err := json.Marshal(&lvl)
	if err != nil {
		return err
	}

	fmt.Println(string(lvlJSON))

	return nil
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

func exit(format string, a ...any) {
	fmt.Fprintf(stdos.Stderr, format+"\n", a...)
	stdos.Exit(1)
}
