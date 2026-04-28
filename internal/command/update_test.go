package command

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matteo-gildone/tokensnap/internal/parser"
)

func TestUpdate_rotateName(t *testing.T) {
	tests := []struct {
		name         string
		fileName     string
		wantFilename string
	}{
		{
			name:         "rotate default file",
			fileName:     "tokens.snapshot.json",
			wantFilename: "tokens.snapshot.20060102.json",
		},
		{
			name:         "rotate custom file",
			fileName:     "diff.snapshot.json",
			wantFilename: "diff.snapshot.20060102.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := time.Date(2006, 01, 02, 0, 0, 0, 0, time.UTC)

			got := rotateName(tt.fileName, date)

			if got != tt.wantFilename {
				t.Errorf("want: %q, got: %q", tt.fileName, got)
			}
		})
	}
}

func TestUpdate_snapshotExists(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		wantSetup bool
		want      bool
	}{
		{
			name:      "default file doesn't exist",
			fileName:  "tokens.snapshot.json",
			wantSetup: false,
			want:      false,
		},
		{
			name:      "default file does exists",
			fileName:  "tokens.snapshot.json",
			wantSetup: true,
			want:      true,
		},
		{
			name:      "custom file doesn't exist",
			fileName:  "diff.snapshot.json",
			wantSetup: false,
			want:      false,
		},
		{
			name:      "custom file does exists",
			fileName:  "diff.snapshot.json",
			wantSetup: true,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, tt.fileName)

			if tt.wantSetup {
				mustWriteFile(t, filePath)
			}

			got := snapshotExists(filePath)

			if got != tt.want {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func TestExecUpdate_noExistingSnapshot(t *testing.T) {
	tempTokensDir, tempSnapshotDir := mustSetupDirs(t)
	tempSnapshotFile := "diff.json"
	tempSnapshotPath := filepath.Join(tempSnapshotDir, tempSnapshotFile)
	want := parser.Snapshot{"color.white": "#fff"}

	mustWriteJSONTokens(t, tempTokensDir, `{"color":{"white":{"$value":"#fff"}}}`)

	err := execUpdate(tempTokensDir, tempSnapshotDir, tempSnapshotFile, "", time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !snapshotExists(filepath.Join(tempSnapshotDir, tempSnapshotFile)) {
		t.Errorf("expected %q to exist", tempSnapshotFile)
	}

	assertSnapshot(t, tempSnapshotPath, want)
}

func TestExecUpdate_rotatesExistingSnapshot(t *testing.T) {
	tempTokensDir, tempSnapshotDir := mustSetupDirs(t)
	tempSnapshotFile := "diff.json"
	tempSnapshotPath := filepath.Join(tempSnapshotDir, tempSnapshotFile)
	now := time.Date(2006, 01, 02, 0, 0, 0, 0, time.UTC)
	want := parser.Snapshot{"color.white": "#fff"}

	mustWriteFile(t, tempSnapshotPath)

	mustWriteJSONTokens(t, tempTokensDir, `{"color":{"white":{"$value":"#fff"}}}`)

	err := execUpdate(tempTokensDir, tempSnapshotDir, tempSnapshotFile, "", now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !snapshotExists(tempSnapshotPath) {
		t.Errorf("expected %q to exist", tempSnapshotFile)
	}

	if !snapshotExists(filepath.Join(tempSnapshotDir, rotateName(tempSnapshotFile, now))) {
		t.Errorf("expected %q to exist", filepath.Join(tempSnapshotDir, rotateName(tempSnapshotFile, now)))
	}

	assertSnapshot(t, tempSnapshotPath, want)
}

func TestExecUpdate_missingTokensDir(t *testing.T) {
	_, snapshotDir := mustSetupDirs(t)
	err := execUpdate(filepath.Join(t.TempDir(), "nonexistent"), snapshotDir, "diff.json", "", time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecUpdate_missingSnapshotDir(t *testing.T) {
	tempTokensDir, _ := mustSetupDirs(t)
	err := execUpdate(tempTokensDir, filepath.Join(t.TempDir(), "nonexistent"), "diff.json", "", time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func mustReadSnapshot(t testing.TB, path string) parser.Snapshot {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}
	var s parser.Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("failed to unmarshal snapshot: %v", err)
	}
	return s
}

func mustWriteFile(t testing.TB, filePath string) {
	t.Helper()

	err := os.WriteFile(filePath, []byte("{}"), 0600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
}

func mustWriteJSONTokens(t testing.TB, tempTokensDir, tokens string) {
	t.Helper()

	filePath := filepath.Join(tempTokensDir, "token.json")

	err := os.WriteFile(filePath, []byte(tokens), 0600)
	if err != nil {
		t.Fatalf("failed to create token file: %v", err)
	}
}

func mustSetupDirs(t testing.TB) (tokensDir, snapshotDir string) {
	tempDir := t.TempDir()
	tempTokensDir := filepath.Join(tempDir, "tokens")
	tempSnapshotDir := filepath.Join(tempDir, ".snapshot")

	if err := os.MkdirAll(tempTokensDir, 0o700); err != nil {
		t.Fatalf("failed to create tempTokensDir directory: %v", err)
	}

	if err := os.MkdirAll(tempSnapshotDir, 0o700); err != nil {
		t.Fatalf("failed to create tempSnapshotDir directory: %v", err)
	}

	return tempTokensDir, tempSnapshotDir
}

func assertSnapshot(t testing.TB, path string, want parser.Snapshot) {
	got := mustReadSnapshot(t, path)

	if len(want) != len(got) {
		t.Errorf("want: %d, got : %d", len(want), len(got))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("key %q: want %q, got %q", k, v, got[k])
		}
	}
}
