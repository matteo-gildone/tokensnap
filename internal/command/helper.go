package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/matteo-gildone/tokensnap/internal/parser"
)

// parseTokens walks fsys from its root, skipping excluded directories,
// and returns a merged Snapshot of all token JSON files found.
func parseTokens(fsys fs.FS, excludeFolders map[string]struct{}) (parser.Snapshot, error) {
	snapshots, err := parser.Snapshots(fsys, ".", excludeFolders)
	if err != nil {
		return nil, err
	}

	return snapshots, nil
}

// buildExcludeSet parses a comma-separated list of folder names into a
// set for use with Snapshots.
func buildExcludeSet(exclude string) map[string]struct{} {
	folders := make(map[string]struct{})
	for _, folder := range strings.Split(exclude, ",") {
		trimmed := strings.TrimSpace(folder)
		if trimmed != "" {
			folders[trimmed] = struct{}{}
		}
	}
	return folders
}

// snapshotExists reports whether the file at snapshotFile exists.
func snapshotExists(snapshotFile string) bool {
	_, err := os.Stat(snapshotFile)
	return !errors.Is(err, os.ErrNotExist)
}

// loadSnapshot reads and decodes a snapshot file from path.
func loadSnapshot(path string) (parser.Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}
	var s parser.Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}
	return s, nil
}
