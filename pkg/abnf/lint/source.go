package lint

import (
	"bytes"
	"slices"
)

// source holds a grammar definition together with the CRLF-normalized form
// the linter parses, and maps offsets in the latter back to the former so
// that findings point into the definition as it was given.
//
// A grammar definition is US-ASCII by construction: RFC 5234 builds
// comments, char-vals and prose-vals out of VCHAR and WSP alone. A byte
// offset is therefore also a character offset, which is what makes the
// column numbers below meaningful without decoding.
type source struct {
	original   []byte
	normalized []byte
	// insertions holds the offsets, in the normalized bytes, of the bytes
	// normalization added, in increasing order.
	insertions []int
	// lineStarts holds the offsets, in the original bytes, at which each
	// line begins.
	lineStarts []int
}

func newSource(original []byte) *source {
	normalized := make([]byte, 0, len(original))
	var insertions []int

	for i, b := range original {
		// A line feed that no carriage return precedes is a bare LF ending.
		if b == '\n' && (i == 0 || original[i-1] != '\r') {
			insertions = append(insertions, len(normalized))
			normalized = append(normalized, '\r')
		}
		normalized = append(normalized, b)
	}

	// RFC 5234 terminates every line, the last one included.
	if len(normalized) != 0 && !bytes.HasSuffix(normalized, []byte("\r\n")) {
		insertions = append(insertions, len(normalized), len(normalized))
		normalized = append(normalized, '\r', '\n')
	}

	lineStarts := []int{0}
	for i, b := range original {
		if b == '\n' {
			lineStarts = append(lineStarts, i+1)
		}
	}

	return &source{
		original:   original,
		normalized: normalized,
		insertions: insertions,
		lineStarts: lineStarts,
	}
}

// position maps an offset in the normalized bytes to a position in the
// definition as given.
func (source *source) position(normalizedOffset int) *Position {
	// Every byte inserted before the offset shifts it by one.
	inserted, _ := slices.BinarySearch(source.insertions, normalizedOffset)
	return source.positionOf(max(normalizedOffset-inserted, 0))
}

// positionOf locates an offset in the definition as given.
func (source *source) positionOf(offset int) *Position {
	// The line holding the offset is the last one starting at or before it.
	line, found := slices.BinarySearch(source.lineStarts, offset)
	if !found {
		line--
	}
	line = max(line, 0)

	return &Position{Offset: offset, Line: line + 1, Column: offset - source.lineStarts[line] + 1}
}

// text returns the normalized bytes between two offsets.
func (source *source) text(start int, end int) string {
	if start < 0 || end > len(source.normalized) || start > end {
		return ""
	}
	return string(source.normalized[start:end])
}

// finding makes a finding covering the normalized bytes between two
// offsets, mapped back onto the definition as given.
func (source *source) finding(ruleId RuleId, start int, end int, message string, replacement *string) *Finding {
	return &Finding{
		RuleId:      ruleId,
		Start:       source.position(start),
		End:         source.position(end),
		Message:     message,
		Replacement: replacement,
	}
}

// originalFinding makes a finding covering the bytes between two offsets of
// the definition as given, for what normalization itself acts on.
func (source *source) originalFinding(
	ruleId RuleId,
	start int,
	end int,
	message string,
	replacement *string,
) *Finding {
	return &Finding{
		RuleId:      ruleId,
		Start:       source.positionOf(start),
		End:         source.positionOf(end),
		Message:     message,
		Replacement: replacement,
	}
}

// replacement returns a pointer to the given replacement text, marking a
// finding the linter knows how to act on.
func replacement(text string) *string {
	return &text
}
