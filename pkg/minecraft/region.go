package minecraft

import (
	"encoding/json"
	"fmt"
	"sync"

	stdos "os"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/level/block"
	"github.com/Tnze/go-mc/save"
)

type Region interface {
	ReadSector(x, z int) (data []byte, err error)
	ExistSector(x, z int) bool
	Close() error
}

type Block struct {
	ID string
}

func ReadRegionsBlocks(r Region, c chan<- Block) {
	// chestEntity := block.ChestEntity{}
	// chestID := block.EntityTypes[chestEntity.ID()]
	defer r.Close()

	// TODO(tauraamui): re-write all of this

	wg := sync.WaitGroup{}
	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			if !r.ExistSector(i, j) {
				continue
			}

			data := must(r.ReadSector(i, j))

			wg.Add(1)
			go func(wg *sync.WaitGroup, data []byte) {
				defer wg.Done()
				/*
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
				*/

				var sc save.Chunk
				sc.Load(data)

				lc := must(level.ChunkFromSave(&sc))

				for i := 0; i < len(lc.BlockEntity); i++ {
					be := lc.BlockEntity[i]
					c <- Block{ID: block.EntityList[be.Type].ID()}
				}

				count := len(lc.Sections)

				if count == 0 {
					return
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

						c <- Block{ID: b.ID()}
					}
				}
			}(&wg, data)
		}
	}

	wg.Wait()
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
