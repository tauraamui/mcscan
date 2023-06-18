package minecraft_test

import (
	"testing"
	"testing/fstest"
)

func buildMockFS() fstest.MapFS {
	return fstest.MapFS{
		"config/minecraft/saves/test world/region/r.0.0.mca": &fstest.MapFile{
			Data: region0,
		},
	}
}

func TestWorldOpenReturnsWorldReferenceWithNoError(t *testing.T) {
	/*
		is := is.New(t)
		world, err := mc.OpenWorld(buildMockFS(), "config/minecraft/saves/test world")
		is.NoErr(err)

		is.True(world != nil)
		is.Equal(world.Name(), "test world")
	*/
}
