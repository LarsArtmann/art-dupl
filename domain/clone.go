// Package domain provides core domain types for clone detection.
package domain

import (
	stderrors "errors"
	"fmt"
)

// Clone represents a code clone with strong typing.
//
// Memory Layout Optimized: 8B fields grouped together, 16B fields at end
// to minimize padding and improve cache locality.
//
// Note: Clone intentionally omits an ID field. Identity is determined by the
// combination of Filename, StartLine, EndLine, and Hash. This eliminates
// unnecessary storage overhead and avoids cargo-cult "entities need IDs" patterns.
type Clone struct {
	StartLine  LineNumber          `json:"startLine"`
	EndLine    LineNumber          `json:"endLine"`
	StartPos   BytePosition        `json:"startPos"`
	EndPos     BytePosition        `json:"endPos"`
	Complexity ComplexityScore     `json:"complexity"`
	Filename   StringID            `json:"filename"`
	Fragment   StringID            `json:"fragment"`
	Hash       StringID            `json:"hash"`
	Status     FileProcessingState `json:"status"`
}

func (c Clone) IsValid() error {
	if c.EndLine < c.StartLine {
		return stderrors.New("clone end line must be >= start line")
	}
	if c.StartPos > 0 || c.EndPos > 0 {
		if c.StartPos >= c.EndPos {
			return stderrors.New("clone end position must be > start position")
		}
	}
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid clone processing state: %s", c.Status)
	}
	return nil
}

func (c Clone) FilenameString() string { return GlobalPool().Lookup(c.Filename) }
func (c Clone) FragmentString() string { return GlobalPool().Lookup(c.Fragment) }
func (c Clone) HashString() string     { return GlobalPool().Lookup(c.Hash) }

func (c *Clone) SetFilename(filename string) { c.Filename = GlobalPool().Intern(filename) }
func (c *Clone) SetFragment(fragment string) { c.Fragment = GlobalPool().Intern(fragment) }
func (c *Clone) SetHash(hash string)         { c.Hash = GlobalPool().Intern(hash) }

// CloneGroup represents a group of clones.
type CloneGroup struct {
	ID       string              `json:"id"`
	Clones   []Clone             `json:"clones"`
	Hash     string              `json:"hash"`
	Size     uint                `json:"size"`
	Severity CloneSeverity       `json:"severity"`
	Status   FileProcessingState `json:"status"`
}

func (cg CloneGroup) IsValid() error {
	if len(cg.Clones) == 0 {
		return stderrors.New("clone group must have at least one clone")
	}
	if !cg.Severity.IsValid() {
		return fmt.Errorf("invalid clone severity: %s", cg.Severity)
	}
	if !cg.Status.IsValid() {
		return fmt.Errorf("invalid clone group status: %s", cg.Status)
	}
	for i, clone := range cg.Clones {
		if err := clone.IsValid(); err != nil {
			return fmt.Errorf("clone %d in group is invalid: %w", i, err)
		}
	}
	return nil
}
