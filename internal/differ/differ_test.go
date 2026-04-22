package differ

import (
	"slices"
	"testing"

	"github.com/matteo-gildone/tokensnap/internal/parser"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		baseline parser.Snapshot
		current  parser.Snapshot
		diff     Diff
	}{
		{
			name: "identical snapshots",
			baseline: parser.Snapshot{
				"color.white": "#fff",
			},
			current: parser.Snapshot{
				"color.white": "#fff",
			},
			diff: Diff{},
		},
		{
			name:     "key added",
			baseline: parser.Snapshot{},
			current: parser.Snapshot{
				"color.white": "#fff",
			},
			diff: Diff{
				Changes: []Change{
					{Key: "color.white", Kind: Added, NewValue: "#fff"},
				},
			},
		},
		{
			name: "key removed",
			baseline: parser.Snapshot{
				"color.white": "#fff",
			},
			current: parser.Snapshot{},
			diff: Diff{
				Changes: []Change{
					{Key: "color.white", Kind: Removed, OldValue: "#fff"},
				},
			},
		},
		{
			name: "value changed",
			baseline: parser.Snapshot{
				"spacing.xl": "16px",
			},
			current: parser.Snapshot{
				"spacing.xl": "24px",
			},
			diff: Diff{
				Changes: []Change{
					{Key: "spacing.xl", Kind: Changed, OldValue: "16px", NewValue: "24px"},
				},
			},
		},
		{
			name: "mixed changes",
			baseline: parser.Snapshot{
				"spacing.xl":  "16px",
				"color.white": "#fff",
			},
			current: parser.Snapshot{
				"spacing.xl": "24px",
			},
			diff: Diff{
				Changes: []Change{
					{Key: "spacing.xl", Kind: Changed, OldValue: "16px", NewValue: "24px"},
					{Key: "color.white", Kind: Removed, OldValue: "#fff"},
				},
			},
		},
		{
			name:     "both empty",
			baseline: parser.Snapshot{},
			current:  parser.Snapshot{},
			diff:     Diff{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compare(tt.baseline, tt.current)

			if !slices.Equal(got.Changes, tt.diff.Changes) {
				t.Errorf("want %v, got: %v", tt.diff.Changes, got.Changes)
			}
		})
	}
}

func TestDiff_IsClean(t *testing.T) {
	tests := []struct {
		name     string
		baseline parser.Snapshot
		current  parser.Snapshot
		diff     Diff
		want     bool
	}{
		{
			name: "has changes",
			baseline: parser.Snapshot{
				"spacing.xl":  "16px",
				"color.white": "#fff",
			},
			current: parser.Snapshot{
				"spacing.xl": "24px",
			},
			want: false,
		},
		{
			name:     "no changes",
			baseline: parser.Snapshot{},
			current:  parser.Snapshot{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := Compare(tt.baseline, tt.current)

			if diff.IsClean() != tt.want {
				t.Errorf("want %v, got: %v", tt.want, diff.IsClean())
			}
		})
	}
}
