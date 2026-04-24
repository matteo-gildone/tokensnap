package differ

import (
	"slices"
	"strings"

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

	slices.SortFunc(changes.Changes, func(a, b Change) int {
		if a.Kind != b.Kind {
			return kindOrder(a.Kind) - kindOrder(b.Kind)
		}
		return strings.Compare(a.Key, b.Key)
	})

	return changes
}

// IsClean returns true if the diff contains no changes.
func (d Diff) IsClean() bool {
	return len(d.Changes) == 0
}

func kindOrder(k ChangeKind) int {
	switch k {
	case Added:
		return 0
	case Changed:
		return 1
	case Removed:
		return 2
	default:
		return 3
	}
}
