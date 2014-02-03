package redux

import "path/filepath"

// RelPath is a structure to simplify paths from
// directory upwards traversal
type RelPath struct {
	entries []string
}

// Add an entry
func (r *RelPath) Add(s string) {
	r.entries = append(r.entries, s)
}

// Reverse and Join
func (r *RelPath) Join() string {
	s := make([]string, len(r.entries))
	for i, j := 0, len(s)-1; i < len(s); i, j = i+1, j-1 {
		s[i] = r.entries[j]
	}
	return filepath.Join(s...)
}
