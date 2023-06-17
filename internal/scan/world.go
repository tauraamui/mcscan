package scan

import (
	"compress/gzip"
	"io/fs"
	"path/filepath"

	"github.com/Tnze/go-mc/save"
)

type World struct {
	fsys    fs.FS
	dirFD   fs.File
	path    string
	name    string
	lvlFD   fs.File
	regions []region
}

type Level struct {
	save.Level
}

func OpenWorld(fsys fs.FS, path string) (*World, error) {
	fd, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}

	w := World{fsys: fsys, dirFD: fd, path: path, name: filepath.Base(path)}

	if err := w.openLevel(); err != nil {
		return nil, err
	}

	if err := w.openRegions(); err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *World) openRegions() error {
	return nil
}

func (w *World) openLevel() error {
	fd, err := w.fsys.Open(filepath.Join(w.path, "level.dat"))
	if err != nil {
		return err
	}

	w.lvlFD = fd
	return nil
}

func (w World) Name() string {
	return w.name
}

func (w World) ReadLevel() (*Level, error) {
	r, err := gzip.NewReader(w.lvlFD)
	if err != nil {
		return nil, err
	}

	lvl, err := save.ReadLevel(r)
	if err != nil {
		return nil, err
	}
	return &Level{lvl}, nil
}

func (w World) WriteLevel(lvl *Level) error {

	return nil
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
