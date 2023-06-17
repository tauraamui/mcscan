package main

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	stdos "os"
	"path/filepath"
	"sync"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/save/region"
	"github.com/dgraph-io/badger/v3"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/os"
	"github.com/tauraamui/mcscan/internal/scan"
	"github.com/tauraamui/mcscan/storage"
	"github.com/tauraamui/mcscan/vfs"
)

type fsResolver func() (*os.FS, error)
type dbResolver func() (storage.DB, error)

func main() {
	rch := make(chan regionBlocks)
	go func() {
		if err := resolveRegionCounts(resolveFS, rch); err != nil {
			panic(err)
		}
	}()

	worldTotalCount := map[string]uint64{}

	for r := range rch {
		for k, v := range r.counts {

			count, ok := worldTotalCount[k]
			if !ok {
				worldTotalCount[k] = v
			}

			worldTotalCount[k] = count + v
		}
	}

	for k, v := range worldTotalCount {
		fmt.Printf("%s %d\n", k, v)
	}

	// storeBlockFrequencies()
	// scanPlayerData()
}

func resolveDB() (storage.DB, error) {
	return storage.NewMemDB()
}

func resolveFS() (*os.FS, error) {
	fs := os.NewFS()

	workingDirectory, _ := stdos.Getwd()                  // Get current working directory
	workingDirectory, _ = fs.FromOSPath(workingDirectory) // Convert to an FS path
	workingDirFS, _ := fs.Sub(workingDirectory)           // Run all file system operations rooted at the current working directory

	ofs, ok := workingDirFS.(*os.FS)
	if !ok {
		return nil, errors.New("sub FS not an OS instance FS")
	}

	return ofs, nil
}

type readerWriterSeeker struct {
	fd hackpadfs.File
}

func (s *readerWriterSeeker) Read(p []byte) (n int, err error) {
	return s.fd.Read(p)
}

func (s *readerWriterSeeker) Write(p []byte) (n int, err error) {
	return hackpadfs.WriteFile(s.fd, p)
}

func (s *readerWriterSeeker) Seek(offset int64, whence int) (int64, error) {
	return hackpadfs.SeekFile(s.fd, offset, whence)
}

type regionBlocks struct {
	size   float64
	name   string
	counts map[string]uint64
}

func resolveRegionCounts(fsr fsResolver, c chan<- regionBlocks) error {
	defer close(c)

	fsys, err := fsr()
	if err != nil {
		return err
	}

	rootpath := filepath.Join("testdata", "region", "*.mca")

	found, err := vfs.Glob(fsys, rootpath)
	if err != nil {
		return err
	}

	wwg := sync.WaitGroup{}
	wwg.Add(len(found))
	for _, f := range found {
		go func(wwg *sync.WaitGroup, f string, c chan<- regionBlocks) {
			defer wwg.Done()
			fd := must(fsys.Open(f))
			defer fd.Close()
			fi := must(fd.Stat())

			rregion := must(region.Load(&readerWriterSeeker{fd: fd}))
			defer rregion.Close()

			totalCounts := map[string]uint64{}
			blocks := make(chan scan.Block)

			wg := sync.WaitGroup{}
			wg.Add(2)

			go func(wg *sync.WaitGroup) {
				defer wg.Done()

				for b := range blocks {
					blockCountKey := b.ID

					count, ok := totalCounts[blockCountKey]
					if !ok {
						totalCounts[blockCountKey] = 1
					}

					totalCounts[blockCountKey] = count + 1
				}
			}(&wg)

			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				scan.Chunks(rregion, blocks)
			}(&wg)

			wg.Wait()

			c <- regionBlocks{
				size:   float64(fi.Size()) / 1024 / 1024,
				name:   f,
				counts: totalCounts,
			}
		}(&wwg, f, c)
	}

	wwg.Wait()

	return nil
}

func storeBlockFrquenciesWithVFS(fsr fsResolver, dbr dbResolver) error {
	fsys, err := fsr()
	if err != nil {
		return err
	}

	rootpath := filepath.Join("testdata", "region", "*.mca")

	found, err := vfs.Glob(fsys, rootpath)
	if err != nil {
		return err
	}

	wwg := sync.WaitGroup{}
	wwg.Add(len(found))
	for _, f := range found {
		go func(wwg *sync.WaitGroup, f string) {
			defer wwg.Done()
			fd := must(fsys.Open(f))
			fi := must(fd.Stat())

			outputHeader := fmt.Sprintf("region file: %s %fMb", f, float64(fi.Size())/1024/1024) // convert bytes to Mb for printout
			rregion := must(region.Load(&readerWriterSeeker{fd: fd}))

			defer rregion.Close()

			totalCounts := map[string]uint64{}

			blocks := make(chan scan.Block)

			wg := sync.WaitGroup{}
			wg.Add(2)

			go func(wg *sync.WaitGroup) {
				defer wg.Done()

				for b := range blocks {
					blockCountKey := b.ID

					count, ok := totalCounts[blockCountKey]
					if !ok {
						totalCounts[blockCountKey] = 1
					}

					totalCounts[blockCountKey] = count + 1
				}
			}(&wg)

			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				scan.Chunks(rregion, blocks)
			}(&wg)

			wg.Wait()

			for k, v := range totalCounts {
				fmt.Printf("%s, KEY: %s, VALUE: %d\n", outputHeader, k, v)
			}
		}(&wwg, f)
	}

	wwg.Wait()

	return nil
}

func storeBlockFrequencies() {
	rootpath := filepath.Join("testdata", "region", "*.mca")

	db := must(storage.NewMemDB())
	defer db.Close()

	fs := must(filepath.Glob(rootpath))
	for _, f := range fs {
		scanChunksSections(f, &db)
	}

	if err := db.DumpToStdout(); err != nil {
		panic(err)
	}
}

type blockEntityTag struct {
	Items []save.Item
}

func (b *blockEntityTag) unmarshal(d any) error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, b)
}

func scanPlayerData() {
	uuid := "480c70ff-1bf6-44e3-8e42-f365f2d4fbef"
	playerDataFD := must(stdos.Open(filepath.Join("testdata", "playerdata", uuid+".dat")))
	defer playerDataFD.Close()

	gReader := must(gzip.NewReader(playerDataFD))

	data := must(save.ReadPlayerData(gReader))

	fmt.Printf("%s's inventory contents:\n", uuid)
	for _, i := range data.Inventory {
		fmt.Printf("    %s\n", i.ID)
		if len(i.Tag) > 0 {
			if beTagData, ok := i.Tag["BlockEntityTag"]; ok {
				beTag := blockEntityTag{}
				beTag.unmarshal(beTagData)
				fmt.Printf("    contains:\n")
				for _, item := range beTag.Items {
					fmt.Printf("        %+v\n", item)
				}
			}
		}
	}

	fmt.Printf("%s's ender chest contents:\n", uuid)
	for _, i := range data.EnderItems {
		fmt.Printf("    %s\n", i.ID)
		if len(i.Tag) > 0 {
			if beTagData, ok := i.Tag["BlockEntityTag"]; ok {
				beTag := blockEntityTag{}
				beTag.unmarshal(beTagData)
				fmt.Printf("    contains:\n")
				for _, item := range beTag.Items {
					fmt.Printf("        %+v\n", item)
				}
			}
		}
	}
}

func scanChunksSections(path string, db *storage.DB) {
	adders := map[string]*badger.MergeOperator{}
	defer func() {
		for _, adder := range adders {
			adder.Stop()
		}
	}()

	r0 := must(region.Open(path))
	defer r0.Close()

	chestEntity := block.ChestEntity{}
	chestID := block.EntityTypes[chestEntity.ID()]

	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			if !r0.ExistSector(i, j) {
				continue
			}

			data := must(r0.ReadSector(i, j))

			var sc save.Chunk
			sc.Load(data)

			lc := must(level.ChunkFromSave(&sc))

			for i := 0; i < len(lc.BlockEntity); i++ {
				be := lc.BlockEntity[i]
				if chestID == be.Type {
					beTagData := blockEntityTag{}
					be.Data.Unmarshal(&beTagData)
					if len(beTagData.Items) > 0 {
						fmt.Printf("player placed chest items: %+v\n", beTagData.Items)
					}
				}
			}

			count := len(lc.Sections)

			if count == 0 {
				continue
			}

			for i := 0; i < count; i++ {
				sec := lc.Sections[i]
				blockCount := int(sec.BlockCount)
				if blockCount == 0 {
					continue
				}

				for j := 0; j < blockCount; j++ {
					b := block.StateList[sec.GetBlock(j)]

					if block.IsAirBlock(b) {
						continue
					}

					blockCountKey := fmt.Sprintf("%s:%s", b.ID(), "frequency")

					adder, ok := adders[blockCountKey]
					if !ok {
						adder = db.Adder([]byte(blockCountKey))
						adders[blockCountKey] = adder
					}

					adder.Add(uint64ToBytes(1))
				}
			}
		}
	}
}

func uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func scanChunksBlockEntities(path string) {
	r0, err := region.Open(path)
	if err != nil {
		panic(err)
	}

	defer r0.Close()

	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			if !r0.ExistSector(i, j) {
				continue
			}

			data := must(r0.ReadSector(i, j))

			var sc save.Chunk
			sc.Load(data)

			lc := must(level.ChunkFromSave(&sc))

			count := len(lc.BlockEntity)
			if count == 0 {
				continue
			}
			fmt.Printf("%s [X/Y/Z]: [%d, %d, %d] - %d item blocks in chunk\n", filepath.Base(path), sc.XPos, sc.YPos, sc.ZPos, count)

			for i := 0; i < count; i++ {
				x, z := lc.BlockEntity[i].UnpackXZ()
				y := int(lc.BlockEntity[i].Y)

				fmt.Printf("\t%d [X/Y/Z]: [%d, %d, %d] TYPE: %v\n", i, x, y, z, block.EntityList[lc.BlockEntity[i].Type].ID())
			}
		}
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		_, _ = fmt.Fprintln(stdos.Stderr, err)
		stdos.Exit(1)
	}
	return v
}
