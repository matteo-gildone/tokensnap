package command

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/matteo-gildone/tokensnap/internal/parser"
)

var UpdateCmd = &Command{
	Name:  "update",
	Usage: usage(),
	Run:   runUpdate,
}

func usage() string {
	return "tokensnap update [-tokens-dir] [-snapshot-file] [-snapshot-dir]"
}

func runUpdate(args []string) error {
	updateSubcommand := flag.NewFlagSet("update", flag.ExitOnError)
	tokensDir := updateSubcommand.String("tokens-dir", "./tokens", "directory containing token JSON files")
	snapshotDir := updateSubcommand.String("snapshot-dir", "./", "directory to write the snapshot file")
	snapshotFile := updateSubcommand.String("snapshot-file", "tokens.snapshot.json", "snapshot filename")
	excludeFolder := updateSubcommand.String("exclude", ".git,vendor,node_modules", "folders to exclude")
	updateSubcommand.Parse(args)

	if updateSubcommand.NArg() != 0 {
		return fmt.Errorf("usage: %s", usage())
	}

	return execUpdate(*tokensDir, *snapshotDir, *snapshotFile, *excludeFolder, time.Now())
}

// execUpdate generates a fresh snapshot from tokensDir, rotates any
// existing snapshot file using now for the datestamp, and writes the
// new snapshot to snapshotDir/snapshotFile.
func execUpdate(tokensDir, snapshotDir, snapshotFile, exclude string, now time.Time) error {
	excludeFoldersDefaults := []string{".git", "vendor", "node_modules", "testdata", "script"}
	excludeFolders := make(map[string]struct{})

	for _, d := range excludeFoldersDefaults {
		excludeFolders[d] = struct{}{}
	}

	for _, folder := range strings.Split(exclude, ",") {
		trimmed := strings.TrimSpace(folder)
		if trimmed != "" {
			excludeFolders[trimmed] = struct{}{}
		}
	}

	snapshotPath := filepath.Join(snapshotDir, snapshotFile)

	if snapshotExists(snapshotPath) {
		err := os.Rename(snapshotPath, rotateName(snapshotPath, now))
		if err != nil {
			return fmt.Errorf("failed to rename file: %w", err)
		}
	}

	fsys := os.DirFS(tokensDir)
	snapshots, err := parseTokens(fsys, excludeFolders)
	if err != nil {
		return err
	}

	return saveSnapshot(snapshotPath, snapshots)
}

// saveSnapshot save snapshot file
func saveSnapshot(snapshotPath string, snapshots parser.Snapshot) error {
	data, err := json.Marshal(snapshots)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshots: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(snapshotPath), "tokens.snapshot-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	removeTemp := true
	defer func() {
		if removeTemp {
			os.Remove(tmp.Name())
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	removeTemp = false
	return os.Rename(tmp.Name(), snapshotPath)
}

// snapshotExists check if a snapshot file is present in the project
func snapshotExists(snapshotFile string) bool {
	_, err := os.Stat(snapshotFile)
	return !errors.Is(err, os.ErrNotExist)
}

// rotateName returns the filename to use when rotating snapshotFile.
// The date is formatted as YYYYMMDD and inserted before the .json extension.
func rotateName(snapshotFile string, t time.Time) string {
	rName := strings.TrimSuffix(snapshotFile, ".json")
	return fmt.Sprintf("%s.%s.json", rName, t.Format("20060102"))
}
