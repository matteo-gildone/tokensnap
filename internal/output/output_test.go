package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/matteo-gildone/tokensnap/internal/differ"
)

func TestTextWriter(t *testing.T) {
	tests := []struct {
		name string
		diff differ.Diff
		want string
	}{
		{
			name: "no changes",
			diff: differ.Diff{},
			want: "no drift detected",
		},
		{
			name: "print header",
			diff: differ.Diff{
				Changes: []differ.Change{
					{Key: "color.white", Kind: differ.Added, NewValue: "#fff"},
					{Key: "color.black", Kind: differ.Removed, OldValue: "#000"},
					{Key: "spacing.xl", Kind: differ.Changed, OldValue: "16px", NewValue: "24px"},
				},
			},
			want: "3 changes detected:",
		},
		{
			name: "add changes",
			diff: differ.Diff{
				Changes: []differ.Change{
					{Key: "color.white", Kind: differ.Added, NewValue: "#fff"},
				},
			},
			want: "+ color.white →  #fff",
		},
		{
			name: "removed changes",
			diff: differ.Diff{
				Changes: []differ.Change{
					{Key: "color.white", Kind: differ.Removed, OldValue: "#fff"},
				},
			},
			want: "- color.white",
		},
		{
			name: "changed changes",
			diff: differ.Diff{
				Changes: []differ.Change{
					{Key: "spacing.xl", Kind: differ.Changed, OldValue: "16px", NewValue: "24px"},
				},
			},
			want: "~ spacing.xl 16px →  24px",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := TextWriter(&buf, tt.diff)
			if err != nil {
				t.Fatal(err)
			}
			output := buf.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("want: %q, got %q", tt.want, output)
			}
		})
	}
}

func TestJSONWriter(t *testing.T) {
	t.Run("no changes", func(t *testing.T) {
		var buf bytes.Buffer
		err := JSONWriter(&buf, differ.Diff{})
		if err != nil {
			t.Fatal(err)
		}

		var r jsonReport
		err = json.Unmarshal(buf.Bytes(), &r)
		if err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if r.Drift {
			t.Errorf("Drift: want false, got true")
		}

		if len(r.Changes) != 0 {
			t.Errorf("expect no changes, got: %d", len(r.Changes))
		}
	})
	t.Run("mixed changes", func(t *testing.T) {
		var buf bytes.Buffer
		err := JSONWriter(&buf, differ.Diff{
			Changes: []differ.Change{
				{Key: "color.white", Kind: differ.Added, NewValue: "#fff"},
				{Key: "spacing.xl", Kind: differ.Changed, OldValue: "16px", NewValue: "24px"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		var r jsonReport
		err = json.Unmarshal(buf.Bytes(), &r)
		if err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if !r.Drift {
			t.Errorf("changes expected")
		}

		if len(r.Changes) != 2 {
			t.Errorf("expect 2 changes")
		}

		if r.Changes[0].Key != "color.white" {
			t.Errorf("want: %q, got: %q", "color.white", r.Changes[0].Key)
		}
		if r.Changes[0].Kind != "added" {
			t.Errorf("want: %q, got: %q", "added", r.Changes[0].Kind)
		}
		if r.Changes[0].New != "#fff" {
			t.Errorf("want: %q, got: %q", "#fff", r.Changes[0].New)
		}

		if r.Changes[1].Key != "spacing.xl" {
			t.Errorf("want: %q, got: %q", "spacing.xl", r.Changes[1].Key)
		}
		if r.Changes[1].Kind != "changed" {
			t.Errorf("want: %q, got: %q", "changed", r.Changes[1].Kind)
		}
		if r.Changes[1].Old != "16px" {
			t.Errorf("want: %q, got: %q", "16px", r.Changes[1].Old)
		}
		if r.Changes[1].New != "24px" {
			t.Errorf("want: %q, got: %q", "24px", r.Changes[1].New)
		}
	})
}
