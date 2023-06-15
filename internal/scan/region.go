package scan

import (
	"encoding/json"
	"fmt"

	stdos "os"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/save"
)

type Region interface {
	ReadSector(x, z int) (data []byte, err error)
	ExistSector(x, z int) bool
}

type FrequenciesByID map[string]uint64

func Chunks(r Region) FrequenciesByID {
	chestEntity := block.ChestEntity{}
	chestID := block.EntityTypes[chestEntity.ID()]

	counts := FrequenciesByID{}

	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			if !r.ExistSector(i, j) {
				continue
			}

			data := must(r.ReadSector(i, j))

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

					blockCountKey := b.ID()

					count, ok := counts[blockCountKey]
					if !ok {
						counts[blockCountKey] = 1
					}

					counts[blockCountKey] = count + 1
				}
			}
		}
	}
	return counts
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

func must[T any](v T, err error) T {
	if err != nil {
		_, _ = fmt.Fprintln(stdos.Stderr, err)
		stdos.Exit(1)
	}
	return v
}
