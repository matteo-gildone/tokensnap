package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

type Snapshot map[string]string

type CollisionError struct {
	Key   string
	FileA string
	FileB string
}

func (e *CollisionError) Error() string {
	return fmt.Sprintf("%q present in %q and %q", e.Key, e.FileA, e.FileB)
}

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

func Snapshots(fsys fs.FS, root string, exclude map[string]struct{}) (Snapshot, error) {
	snapshots := make(Snapshot)
	visited := make(map[string]string)
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if _, ok := exclude[d.Name()]; ok {
				return fs.SkipDir
			}
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		if filepath.Ext(path) != ".json" {
			return nil
		}

		f, err := fsys.Open(path)
		if err != nil {
			return fmt.Errorf("open %s: %w", path, err)
		}
		defer f.Close()

		ss, err := ParseFile(f)

		if err != nil {
			return fmt.Errorf("failed to parse content: %w", err)
		}

		for k, value := range ss {
			if v, ok := visited[k]; ok {
				return &CollisionError{
					Key:   k,
					FileA: v,
					FileB: path,
				}
			}
			snapshots[k] = value
			visited[k] = path
		}

		return nil
	})
	return snapshots, err
}
