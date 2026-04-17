package parser

import (
	"encoding/json"
	"fmt"
	"io"
)

type Snapshot map[string]string

// ParseFile reads the JSON file at path, flattens it according to the
// W3C Design Token rules, and returns the result as a Snapshot.
// $type and name fields are skipped. $value fields are emitted as-is,
// including alias references such as "{color.white}".
func ParseFile(r io.Reader) (Snapshot, error) {
	var raw map[string]any
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed decode json: %w", err)
	}

	tokens := make(Snapshot)
	err := flatten(raw, "", tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// flatten is the recursive core. It walks a decoded JSON value
// (map[string]any from json.Unmarshal), building the dot-path as it
// descends, and writes leaf token entries into dst.
// It is unexported — ParseFile is the public entry point.
func flatten(node map[string]any, prefix string, dst Snapshot) error {
	for k, v := range node {
		if k == "$type" || k == "$description" {
			continue
		}

		childPath := k
		if prefix != "" {
			childPath = prefix + "." + k
		}

		m, ok := v.(map[string]any)
		if !ok {
			continue
		}

		if value, ok := m["$value"]; ok {
			dst[childPath] = fmt.Sprint(value)
		} else {
			err := flatten(m, childPath, dst)
			if err != nil {
				return fmt.Errorf("failed flatten: %w", err)
			}
		}
	}
	return nil
}
