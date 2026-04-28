package command

import (
	"io/fs"

	"github.com/matteo-gildone/tokensnap/internal/parser"
)

func parseTokens(fsys fs.FS, excludeFolders map[string]struct{}) (parser.Snapshot, error) {
	snapshots, err := parser.Snapshots(fsys, ".", excludeFolders)
	if err != nil {
		return nil, err
	}

	return snapshots, nil
}
