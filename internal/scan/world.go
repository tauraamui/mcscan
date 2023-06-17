package scan

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/save"
	"github.com/tauraamui/mcscan/internal/filesystem"
)

type World struct {
	fsys    filesystem.FS
	dirFD   fs.File
	path    string
	name    string
	lvlFD   fs.File
	regions []region
}

type Level struct {
	save.Level
}

func OpenWorld(fsys filesystem.FS, path string) (*World, error) {
	fd, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}

	w := World{fsys: fsys, dirFD: fd, path: path, name: filepath.Base(path)}

	if err := w.openRegions(); err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *World) openRegions() error {
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

func (w World) CountBlocks() {}

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

type region struct {
	path string
}
