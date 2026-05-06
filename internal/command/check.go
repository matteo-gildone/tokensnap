package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/matteo-gildone/tokensnap/internal/differ"
	"github.com/matteo-gildone/tokensnap/internal/output"
)

var ErrDrift = errors.New("changes available")

type checkConfig struct {
	TokenDir      string
	SnapshotDir   string
	SnapshotFile  string
	ExcludeFolder string
}

func (c checkConfig) snapshotPath() string {
	return filepath.Join(c.SnapshotDir, c.SnapshotFile)
}

var CheckCmd = &Command{
	Name:  "check",
	Run:   runCheck,
	Usage: checkUsage(),
}

func checkUsage() string {
	return "tokensnap check [-tokens-dir] [-snapshot-file] [-snapshot-dir] [-format]"
}

func runCheck(args []string) error {
	checkCommand := flag.NewFlagSet("check", flag.ExitOnError)
	tokensDir := checkCommand.String("tokens-dir", "./tokens", "directory containing token JSON files")
	snapshotDir := checkCommand.String("snapshot-dir", "./", "directory to write the snapshot file")
	snapshotFile := checkCommand.String("snapshot-file", "tokens.snapshot.json", "snapshot filename")
	format := checkCommand.String("format", "text", "output format, text or json")
	excludeFolder := checkCommand.String("exclude", ".git,vendor,node_modules", "folders to exclude")
	checkCommand.Parse(args)
	if checkCommand.NArg() != 0 {
		return fmt.Errorf("usage: %s", checkUsage())
	}
	cfg := checkConfig{TokenDir: *tokensDir, SnapshotDir: *snapshotDir, SnapshotFile: *snapshotFile, ExcludeFolder: *excludeFolder}
	return execCheck(cfg, *format, os.Stdout)
}

// execCheck diffs current tokens against the saved snapshot and writes
// output to w. Returns error.
func execCheck(cfg checkConfig, format string, out io.Writer) error {
	excludeFolders := buildExcludeSet(cfg.ExcludeFolder)
	snapshotFile := cfg.snapshotPath()

	if !snapshotExists(snapshotFile) {
		return fmt.Errorf("snapshot not found at %q: run tokensnap update first", snapshotFile)
	}

	baseline, err := loadSnapshot(snapshotFile)
	if err != nil {
		return err
	}

	fsys := os.DirFS(cfg.TokenDir)
	snapshots, err := parseTokens(fsys, excludeFolders)
	if err != nil {
		return err
	}

	d := differ.Compare(baseline, snapshots)

	switch format {
	case "json":
		err = output.JSONWriter(out, d)
	default:
		err = output.TextWriter(out, d)
	}
	if err != nil {
		return err
	}

	if !d.IsClean() {
		return ErrDrift
	}
	return nil
}
