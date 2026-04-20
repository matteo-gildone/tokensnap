package parser

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    Snapshot
		wantErr bool
	}{
		{
			name:  "valid simple token",
			input: `{"color":{"white":{"$value":"#fff"}}}`,
			want: Snapshot{
				"color.white": "#fff",
			},
		},
		{
			name:    "invalid JSON",
			input:   `{"color":{`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error got nil")
					return
				}
			}

			if !maps.Equal(got, tt.want) {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func TestFlatten(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		input        string
		wantSnapshot Snapshot
	}{
		{
			name:  "single token",
			input: `{"color":{"white":{"$value":"#fff"}}}`,
			wantSnapshot: Snapshot{
				"color.white": "#fff",
			},
		},
		{
			name:  "skips $type at root",
			input: `{"$type":"color","color":{"white":{"$value":"#fff"}}}`,
			wantSnapshot: Snapshot{
				"color.white": "#fff",
			},
		},
		{
			name:  "skips name field",
			input: `{"color":{"white":{"name":"white","$value":"#fff"}}}`,
			wantSnapshot: Snapshot{
				"color.white": "#fff",
			},
		},
		{
			name:  "alias value kept as-is",
			input: `{"palette":{"bg":{"$value":"{color.white}"}}}`,
			wantSnapshot: Snapshot{
				"palette.bg": "{color.white}",
			},
		},
		{
			name:  "deeply nested path",
			input: `{"palette":{"bg":{"primary":{"$value":"{color.white}"}}}}`,
			wantSnapshot: Snapshot{
				"palette.bg.primary": "{color.white}",
			},
		},
		{
			name:  "numeric $value",
			input: `{"size":{"base":{"$value":16}}}`,
			wantSnapshot: Snapshot{
				"size.base": "16",
			},
		},
		{
			name:  "boolean $value",
			input: `{"flag":{"on":{"$value":true}}}`,
			wantSnapshot: Snapshot{
				"flag.on": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw map[string]any
			if err := json.Unmarshal([]byte(tt.input), &raw); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			result := make(Snapshot)
			err := flatten(raw, "", result)
			if err != nil {
				t.Fatalf("flatten failed: %v", err)
			}
			if !maps.Equal(result, tt.wantSnapshot) {
				t.Errorf("want: %v, got: %v", tt.wantSnapshot, result)
			}
		})

	}
}
