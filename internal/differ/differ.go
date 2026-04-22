package differ

import (
	"github.com/matteo-gildone/tokensnap/internal/parser"
)

type ChangeKind string

const (
	Added   ChangeKind = "added"
	Removed ChangeKind = "removed"
	Changed ChangeKind = "changed"
)

// Change represents a single diffed token.
type Change struct {
	Key      string
	Kind     ChangeKind
	OldValue string // empty for Added
	NewValue string // empty for Removed
}

// Diff is the full result of comparing two snapshots.
type Diff struct {
	Changes []Change
}

// Filter specifies which ChangeKind values to include in results.
// A nil or empty Filter means include all kinds.
type Filter map[ChangeKind]bool

// Compare returns a Diff describing what changed between baseline
// and current. Changes are sorted: Added first, then Changed, then
// Removed; within each group, sorted alphabetically by key.
func Compare(baseline, current parser.Snapshot) Diff {
	changes := Diff{
		Changes: make([]Change, 0, max(len(baseline), len(current))),
	}

	for k, vb := range baseline {
		vc, ok := current[k]
		if !ok {
			changes.Changes = append(changes.Changes, Change{Key: k, Kind: Removed, OldValue: vb})
		} else if vb != vc {
			changes.Changes = append(changes.Changes, Change{Key: k, Kind: Changed, OldValue: vb, NewValue: vc})
		}
	}

	for k, vc := range current {
		_, ok := baseline[k]
		if !ok {
			changes.Changes = append(changes.Changes, Change{Key: k, Kind: Added, NewValue: vc})
		}
	}

	return changes
}

// IsClean returns true if the diff contains no changes.
func (d Diff) IsClean() bool {
	return len(d.Changes) == 0
}

// Filter returns a new Diff containing only changes
// matching the given Filter. A nil filter returns a copy of d.
//func (d Diff) Filter(f Filter) Diff {}

// ParseFilter parses a comma-separated filter string (e.g.
// "added,changed") into a Filter. Returns an error if any token is
// not a recognised ChangeKind.
//func ParseFilter(s string) (Filter, error) {}
