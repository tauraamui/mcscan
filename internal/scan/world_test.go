package scan_test

import (
	"testing"
	"testing/fstest"

	"github.com/matryer/is"
	"github.com/tauraamui/mcscan/internal/scan"
)

func buildMockFS() fstest.MapFS {
	return fstest.MapFS{
		"config/minecraft/saves/test world/region/r.0.0.mca": &fstest.MapFile{
			Data: []byte{},
		},
	}
}

func TestWorldOpenReturnsWorldReferenceWithNoError(t *testing.T) {
	is := is.New(t)
	world, err := scan.OpenWorld(buildMockFS(), "config/minecraft/saves/test world")
	is.NoErr(err)

	is.True(world != nil)
	is.Equal(world.Name(), "test world")
}
