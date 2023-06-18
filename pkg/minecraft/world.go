package minecraft

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	stdos "os"
	"path/filepath"
	"strings"

	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/save"
	"github.com/hack-pad/hackpadfs"
	"github.com/tauraamui/mcscan/internal/filesystem"
	"github.com/tauraamui/mcscan/internal/vfs"
)

type World struct {
	fsys    filesystem.FS
	dirFD   fs.File
	path    string
	name    string
	lvlFD   fs.File
	regions []region
}

type region struct {
	fsys filesystem.FS
	fd   fs.File
	path string
}

func (s *region) Read(p []byte) (n int, err error) {
	return s.fd.Read(p)
}

func (s *region) Write(p []byte) (n int, err error) {
	return len(p), s.fsys.WriteFile(s.path, p, fs.ModePerm)
}

func (s *region) Seek(offset int64, whence int) (int64, error) {
	return hackpadfs.SeekFile(s.fd, offset, whence)
}

func (r region) blockCount() (map[string]uint64, error) {
	return nil, nil
}

type Level struct {
	save.Level
}

type WorldResolver func(fsys filesystem.FS, ref string) (*World, error)

func OpenWorldByName(fsys filesystem.FS, name string) (*World, error) {
	configDirPath := must(stdos.UserConfigDir())
	configDirPath = strings.TrimPrefix(configDirPath, string(filepath.Separator))
	worldSaveDirPath := filepath.Join(configDirPath, "minecraft", "saves", name)

	fi, err := fsys.Stat(worldSaveDirPath)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("found %s but is not directory", worldSaveDirPath)
	}

	return OpenWorld(fsys, worldSaveDirPath)
}

func OpenWorld(fsys filesystem.FS, path string) (*World, error) {
	fd, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}

	w := World{fsys: fsys, dirFD: fd, path: path, name: filepath.Base(path)}

	if err := w.resolveRegions(); err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *World) resolveRegions() error {
	regionsDir := filepath.Join(w.path, "region")

	found, err := vfs.Glob(w.fsys, regionsDir)
	if err != nil {
		return err
	}

	for _, f := range found {
		w.regions = append(w.regions, region{path: f})
	}

	return nil
}

func (w World) Name() string {
	return w.name
}

func (w World) ReadLevel() (*Level, error) {
	fd, err := w.fsys.Open(filepath.Join(w.path, "level.dat"))
	if err != nil {
		return nil, fmt.Errorf("unable to open level.dat: %w", err)
	}
	defer fd.Close()

	r, err := gzip.NewReader(fd)
	if err != nil {
		return nil, fmt.Errorf("unable to init gzip reader on level.dat file descriptor: %w", err)
	}

	lvl, err := save.ReadLevel(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read level.dat NBT data: %w", err)
	}
	return &Level{lvl}, nil
}

func (w World) WriteLevel(lvl *Level) error {
	data, err := nbt.Marshal(lvl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return err
	}
	gw.Close()

	return w.fsys.WriteFile(filepath.Join(w.path, "level.dat"), buf.Bytes(), fs.ModePerm)
}

func (w World) RegionsCount() int {
	return len(w.regions)
}

func (w World) BlocksCount() {}

func (w World) Close() error {
	// TODO(tauraamui): Should handle errors as independant close failures,
	// and append each error occurance to an errgroup to return at end.
	if w.dirFD != nil {
		if err := w.dirFD.Close(); err != nil {
			return err
		}
	}

	if w.lvlFD != nil {
		if err := w.lvlFD.Close(); err != nil {
			return err
		}
	}

	return nil
}
