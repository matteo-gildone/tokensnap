package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/matteo-gildone/tokensnap/internal/differ"
)

type jsonChange struct {
	Key  string `json:"key"`
	Kind string `json:"kind"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

type jsonReport struct {
	Drift   bool         `json:"drift"`
	Changes []jsonChange `json:"changes"`
}

// TextWriter writes a human-readable diff report to w. Changes are
// grouped by kind (added, changed, removed) with a sigil prefix per line.
func TextWriter(w io.Writer, d differ.Diff) error {
	if d.IsClean() {
		fmt.Fprint(w, "no drift detected\n")
		return nil
	}
	fmt.Fprintf(w, "%d changes detected:\n\n", len(d.Changes))
	var currentKind differ.ChangeKind
	for _, change := range d.Changes {
		if change.Kind != currentKind {
			currentKind = change.Kind
			fmt.Fprintf(w, "%s:\n\n", currentKind)
		}
		fmt.Fprintf(w, "  %s\n", changeOutput(change))
	}
	return nil
}

// JSONWriter writes a machine-readable diff report to w as JSON.
// The output has the shape {"drift": bool, "changes": [...]}.
func JSONWriter(w io.Writer, d differ.Diff) error {
	changes := make([]jsonChange, 0)
	for _, change := range d.Changes {
		changes = append(changes, jsonChange{Key: change.Key, Kind: string(change.Kind), Old: change.OldValue, New: change.NewValue})
	}

	r := &jsonReport{
		Drift:   !d.IsClean(),
		Changes: changes,
	}

	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal json report: %w", err)
	}

	_, err = w.Write(data)
	return err
}

// changeOutput formats a single Change as a human-readable string
// with a sigil prefix: + for added, ~ for changed, - for removed.
func changeOutput(change differ.Change) string {
	switch change.Kind {
	case differ.Added:
		return fmt.Sprintf("+ %s →  %s", change.Key, change.NewValue)
	case differ.Removed:
		return fmt.Sprintf("- %s", change.Key)
	case differ.Changed:
		return fmt.Sprintf("~ %s %s →  %s", change.Key, change.OldValue, change.NewValue)
	default:
		return "unknown change.Kind"
	}
}
