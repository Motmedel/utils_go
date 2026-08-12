package endpoint

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
)

// hasCaseInsensitivePrefix reports whether data starts with the ASCII prefix,
// ignoring case.
func hasCaseInsensitivePrefix(data []byte, prefix string) bool {
	if len(data) < len(prefix) {
		return false
	}
	return bytes.EqualFold(data[:len(prefix)], []byte(prefix))
}

func isHtmlWhitespace(character byte) bool {
	switch character {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

// scanScriptOpenTag scans a script open tag starting at the byte after
// "<script", returning the index just past the closing ">" and whether the tag
// has a src attribute. Quoted attribute values may contain ">".
func scanScriptOpenTag(data []byte, start int) (int, bool) {
	hasSrcAttribute := false

	i := start
	for i < len(data) {
		character := data[i]

		if isHtmlWhitespace(character) || character == '/' {
			i++
			continue
		}
		if character == '>' {
			return i + 1, hasSrcAttribute
		}

		nameStart := i
		for i < len(data) && data[i] != '=' && data[i] != '>' && data[i] != '/' && !isHtmlWhitespace(data[i]) {
			i++
		}
		if bytes.EqualFold(data[nameStart:i], []byte("src")) {
			hasSrcAttribute = true
		}

		for i < len(data) && isHtmlWhitespace(data[i]) {
			i++
		}
		if i < len(data) && data[i] == '=' {
			i++
			for i < len(data) && isHtmlWhitespace(data[i]) {
				i++
			}
			if i < len(data) && (data[i] == '"' || data[i] == '\'') {
				quote := data[i]
				i++
				for i < len(data) && data[i] != quote {
					i++
				}
				i++
			} else {
				for i < len(data) && data[i] != '>' && !isHtmlWhitespace(data[i]) {
					i++
				}
			}
		}
	}

	return i, hasSrcAttribute
}

// extractInlineScripts returns the text contents of inline script elements in
// the provided HTML. It is a scoped scanner intended for build-produced HTML:
// comments are skipped, script contents run to the first case-insensitive
// closing tag (matching raw text element parsing), and script elements with a
// src attribute are ignored.
func extractInlineScripts(data []byte) []string {
	var scripts []string

	i := 0
	for i < len(data) {
		if data[i] != '<' {
			i++
			continue
		}

		if hasCaseInsensitivePrefix(data[i:], "<!--") {
			commentEnd := bytes.Index(data[i+4:], []byte("-->"))
			if commentEnd == -1 {
				break
			}
			i += 4 + commentEnd + 3
			continue
		}

		if !hasCaseInsensitivePrefix(data[i:], "<script") {
			i++
			continue
		}
		boundaryIndex := i + len("<script")
		if boundaryIndex < len(data) && !isHtmlWhitespace(data[boundaryIndex]) &&
			data[boundaryIndex] != '>' && data[boundaryIndex] != '/' {
			i++
			continue
		}

		contentStart, hasSrcAttribute := scanScriptOpenTag(data, boundaryIndex)

		contentEnd := len(data)
		closeIndex := -1
		for j := contentStart; j+len("</script") <= len(data); j++ {
			if hasCaseInsensitivePrefix(data[j:], "</script") {
				closeIndex = j
				break
			}
		}
		if closeIndex != -1 {
			contentEnd = closeIndex
		}

		if !hasSrcAttribute {
			scripts = append(scripts, string(data[contentStart:contentEnd]))
		}

		if closeIndex == -1 {
			break
		}
		i = closeIndex + len("</script")
		for i < len(data) && data[i] != '>' {
			i++
		}
		i++
	}

	return scripts
}

// makeInlineScriptHashes returns deduplicated Content Security Policy hash
// sources for the inline scripts in the provided HTML.
func makeInlineScriptHashes(data []byte) []string {
	var hashes []string
	seen := make(map[string]struct{})

	for _, script := range extractInlineScripts(data) {
		digest := sha256.Sum256([]byte(script))
		hash := "sha256-" + base64.StdEncoding.EncodeToString(digest[:])
		if _, ok := seen[hash]; ok {
			continue
		}
		seen[hash] = struct{}{}
		hashes = append(hashes, hash)
	}

	return hashes
}
