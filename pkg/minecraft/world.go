package minecraft

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	stdos "os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/save"
	mcregion "github.com/Tnze/go-mc/save/region"
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

func (r *region) Read(p []byte) (n int, err error) {
	return r.fd.Read(p)
}

func (r *region) Write(p []byte) (n int, err error) {
	return len(p), r.fsys.WriteFile(r.path, p, fs.ModePerm)
}

func (r *region) Seek(offset int64, whence int) (int64, error) {
	return hackpadfs.SeekFile(r.fd, offset, whence)
}

func (r *region) Close() error {
	return r.fd.Close()
}

type Level struct {
	save.LevelData
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
	regionFiles := filepath.Join(w.path, "region", "*.mca")

	found, err := vfs.Glob(w.fsys, regionFiles)
	if err != nil {
		return err
	}

	for _, f := range found {
		w.regions = append(w.regions, region{fsys: w.fsys, path: f})
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
	return &Level{lvl.Data}, nil
}

func (w World) WriteLevel(lvl *Level) error {
	lll := save.Level{Data: save.LevelData(lvl.LevelData)}
	data, err := nbt.Marshal(lll)
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

func (w World) BlocksCount() (map[string]uint64, error) {
	count := map[string]uint64{}
	if len(w.regions) == 0 {
		return count, nil
	}

	blocks := make(chan Block)
	wg := sync.WaitGroup{}
	for i := 0; i < len(w.regions); i++ {
		rref := w.regions[i]
		fd, err := w.fsys.Open(rref.path)
		if err != nil {
			return nil, err
		}
		rref.fd = fd

		loadedRegion, err := mcregion.Load(&rref)
		if err != nil {
			return nil, err
		}

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			ReadRegionsBlocks(loadedRegion, blocks)
		}(&wg)
	}

	go func(wg *sync.WaitGroup, bc chan Block) {
		defer close(bc)
		wg.Wait()
	}(&wg, blocks)

	for blk := range blocks {
		existingCount, ok := count[blk.ID]
		if ok {
			count[blk.ID] = existingCount + 1
			continue
		}
		count[blk.ID] = 1
	}

	return count, nil
}

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
