package scan

import (
	"io/fs"
	"path/filepath"
)

type FSys interface {
	Open(name string) (fs.File, error)
	fs.StatFS
}

type World struct {
	fsys    FSys
	fd      fs.File
	name    string
	regions []region
}

func OpenWorld(fsys FSys, path string) (*World, error) {
	fd, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}

	w := World{fd: fd, name: filepath.Base(path)}
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

func (w World) CountBlocks() {}

func (w World) Close() error {
	return w.fd.Close()
}

type region struct {
	path string
}
