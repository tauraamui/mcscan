package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/save/region"
)

func main() {
	rootpath := filepath.Join("testdata", "region", "*.mca")

	fs := must(filepath.Glob(rootpath))
	for _, f := range fs {
		scan(f)
	}

}

func scan(path string) {
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
			fmt.Printf("%s - %d item blocks in chunk\n", filepath.Base(path), count)

			for i := 0; i < count; i++ {
				x, z := lc.BlockEntity[i].UnpackXZ()
				y := lc.BlockEntity[i].Y

				fmt.Printf("\t%d [X/Y/Z]: [%d, %d, %d] TYPE: %v\n", i, x, y, z, block.EntityList[lc.BlockEntity[i].Type].ID())
			}
		}
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return v
}
