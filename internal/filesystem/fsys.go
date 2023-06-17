package filesystem

import (
	"io/fs"

	"github.com/hack-pad/hackpadfs"
)

type FS interface {
	Open(name string) (fs.File, error)
	fs.StatFS
	fs.ReadDirFS
	hackpadfs.WriteFileFS
}
