package command

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecCheck_missingSnapshot(t *testing.T) {
	tokensDir, snapshotDir := mustSetupDirs(t)
	cfg := checkConfig{
		TokenDir:      tokensDir,
		SnapshotDir:   snapshotDir,
		SnapshotFile:  "tokens.snapshot.json",
		ExcludeFolder: ".git,vendor,node_modules",
	}
	err := execCheck(cfg, "text", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
func TestExecCheck_noDrift(t *testing.T) {
	tokensDir, snapshotDir := mustSetupDirs(t)
	snapshotFile := "tokens.snapshot.json"
	snapshotPath := filepath.Join(snapshotDir, snapshotFile)
	tokens := `{"color":{"white":{"$value":"#fff"}}}`
	mustWriteJSONTokens(t, tokensDir, tokens)
	cfg := updateConfig{TokenDir: tokensDir, SnapshotDir: snapshotDir, SnapshotFile: snapshotFile, ExcludeFolder: ".git,vendor,node_modules"}
	if err := execUpdate(cfg, fixedTime()); err != nil {
		t.Fatalf("setup: execUpdate failed: %v", err)
	}
	_ = snapshotPath
	var out bytes.Buffer
	checkCfg := checkConfig{TokenDir: tokensDir, SnapshotDir: snapshotDir, SnapshotFile: snapshotFile, ExcludeFolder: ".git,vendor,node_modules"}
	err := execCheck(checkCfg, "text", &out)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out.String(), "no drift detected") {
		t.Errorf("expected %q in output, got %q", "no drift detected", out.String())
	}
}
func TestExecCheck_drift(t *testing.T) {
	tokensDir, snapshotDir := mustSetupDirs(t)
	snapshotFile := "tokens.snapshot.json"
	// write snapshot with old value
	mustWriteJSONTokens(t, tokensDir, `{"color":{"primary":{"$value":"#0057FF"}}}`)
	cfg := updateConfig{TokenDir: tokensDir, SnapshotDir: snapshotDir, SnapshotFile: snapshotFile, ExcludeFolder: ".git,vendor,node_modules"}
	if err := execUpdate(cfg, fixedTime()); err != nil {
		t.Fatalf("setup: execUpdate failed: %v", err)
	}
	// update token to new value
	mustWriteJSONTokens(t, tokensDir, `{"color":{"primary":{"$value":"#0062FF"}}}`)
	var out bytes.Buffer
	checkCfg := checkConfig{TokenDir: tokensDir, SnapshotDir: snapshotDir, SnapshotFile: snapshotFile, ExcludeFolder: ".git,vendor,node_modules"}
	err := execCheck(checkCfg, "text", &out)
	if !errors.Is(err, ErrDrift) {
		t.Fatalf("expected ErrDrift, got %v", err)
	}
	if !strings.Contains(out.String(), "color.primary") {
		t.Errorf("expected diff output to mention %q, got %q", "color.primary", out.String())
	}
}
func TestExecCheck_jsonFormat(t *testing.T) {
	tokensDir, snapshotDir := mustSetupDirs(t)
	snapshotFile := "tokens.snapshot.json"
	mustWriteJSONTokens(t, tokensDir, `{"color":{"primary":{"$value":"#0057FF"}}}`)
	cfg := updateConfig{TokenDir: tokensDir, SnapshotDir: snapshotDir, SnapshotFile: snapshotFile, ExcludeFolder: ".git,vendor,node_modules"}
	if err := execUpdate(cfg, fixedTime()); err != nil {
		t.Fatalf("setup: execUpdate failed: %v", err)
	}
	mustWriteJSONTokens(t, tokensDir, `{"color":{"primary":{"$value":"#0062FF"}}}`)
	var out bytes.Buffer
	checkCfg := checkConfig{TokenDir: tokensDir, SnapshotDir: snapshotDir, SnapshotFile: snapshotFile, ExcludeFolder: ".git,vendor,node_modules"}
	err := execCheck(checkCfg, "json", &out)
	if !errors.Is(err, ErrDrift) {
		t.Fatalf("expected ErrDrift, got %v", err)
	}
	if !strings.Contains(out.String(), `"drift":true`) {
		t.Errorf("expected JSON with drift:true, got %q", out.String())
	}
}

func fixedTime() time.Time {
	return time.Date(2006, 01, 02, 0, 0, 0, 0, time.UTC)
}
