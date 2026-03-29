package mods

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

// ParseDescriptor reads a descriptor.mod file and fills mod metadata fields.
// Unknown keys are silently ignored.
func ParseDescriptor(path string) (name, version, description string, tags []string, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("read descriptor %q: %w", path, err)
	}

	normalized := normalizeDescriptorBytes(content)

	if strings.EqualFold(filepath.Ext(path), ".json") {
		name, version, description, tags, err = parseJSONDescriptor(normalized)
		if err != nil {
			return "", "", "", nil, fmt.Errorf("parse json descriptor %q: %w", path, err)
		}
		return name, version, description, tags, nil
	}

	name, version, description, tags, err = parseTextDescriptor(string(normalized))
	if err != nil {
		return "", "", "", nil, fmt.Errorf("parse text descriptor %q: %w", path, err)
	}

	return name, version, description, tags, nil
}

func parseJSONDescriptor(content []byte) (name, version, description string, tags []string, err error) {
	var payload struct {
		Name             string   `json:"name"`
		Version          string   `json:"version"`
		ShortDescription string   `json:"short_description"`
		Description      string   `json:"description"`
		Tags             []string `json:"tags"`
	}

	if err := json.Unmarshal(content, &payload); err != nil {
		sanitized := sanitizeBrokenJSONStringLiterals(content)
		if bytes.Equal(sanitized, content) {
			return "", "", "", nil, fmt.Errorf("unmarshal descriptor json: %w", err)
		}

		if sanitizeErr := json.Unmarshal(sanitized, &payload); sanitizeErr != nil {
			return "", "", "", nil, fmt.Errorf("unmarshal descriptor json: %w", err)
		}
	}

	description = payload.Description
	if description == "" {
		description = payload.ShortDescription
	}

	return payload.Name, payload.Version, description, payload.Tags, nil
}

func sanitizeBrokenJSONStringLiterals(content []byte) []byte {
	output := make([]byte, 0, len(content)+32)
	inString := false
	escaped := false
	changed := false

	for _, b := range content {
		if !inString {
			output = append(output, b)
			if b == '"' {
				inString = true
			}
			continue
		}

		if escaped {
			output = append(output, b)
			escaped = false
			continue
		}

		switch b {
		case '\\':
			output = append(output, b)
			escaped = true
		case '"':
			output = append(output, b)
			inString = false
		case '\n':
			output = append(output, '\\', 'n')
			changed = true
		case '\r':
			output = append(output, '\\', 'r')
			changed = true
		case '\t':
			output = append(output, '\\', 't')
			changed = true
		default:
			if b < 0x20 {
				output = append(output, '\\', 'u', '0', '0', hexDigit(b>>4), hexDigit(b&0x0F))
				changed = true
				continue
			}
			output = append(output, b)
		}
	}

	if !changed {
		return content
	}

	return output
}

func hexDigit(nibble byte) byte {
	if nibble < 10 {
		return '0' + nibble
	}
	return 'a' + (nibble - 10)
}

func parseTextDescriptor(content string) (name, version, description string, tags []string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "tags") {
			parsedTags, parseErr := parseTagsBlock(line, scanner)
			if parseErr != nil {
				return "", "", "", nil, fmt.Errorf("parse tags block: %w", parseErr)
			}
			tags = parsedTags
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = trimQuoted(value)

		switch key {
		case "name":
			name = value
		case "version":
			version = value
		case "description", "short_description":
			description = value
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return "", "", "", nil, fmt.Errorf("scan descriptor lines: %w", scanErr)
	}

	return name, version, description, tags, nil
}

func normalizeDescriptorBytes(content []byte) []byte {
	if bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}) {
		return content[3:]
	}

	if len(content) >= 2 && (bytes.HasPrefix(content, []byte{0xFF, 0xFE}) || bytes.HasPrefix(content, []byte{0xFE, 0xFF})) {
		return decodeUTF16Descriptor(content)
	}

	return content
}

func decodeUTF16Descriptor(content []byte) []byte {
	if len(content) < 4 {
		return content
	}

	isLittleEndian := content[0] == 0xFF && content[1] == 0xFE
	raw := content[2:]
	if len(raw)%2 != 0 {
		raw = raw[:len(raw)-1]
	}

	units := make([]uint16, 0, len(raw)/2)
	for i := 0; i < len(raw); i += 2 {
		if isLittleEndian {
			units = append(units, uint16(raw[i])|uint16(raw[i+1])<<8)
		} else {
			units = append(units, uint16(raw[i])<<8|uint16(raw[i+1]))
		}
	}

	decoded := utf16.Decode(units)
	return []byte(string(decoded))
}

func parseTagsBlock(firstLine string, scanner *bufio.Scanner) ([]string, error) {
	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(firstLine))

	for !strings.Contains(builder.String(), "}") {
		if !scanner.Scan() {
			if scanErr := scanner.Err(); scanErr != nil {
				return nil, scanErr
			}
			return nil, fmt.Errorf("unterminated tags block")
		}

		builder.WriteString("\n")
		builder.WriteString(strings.TrimSpace(scanner.Text()))
	}

	line := builder.String()
	openIdx := strings.Index(line, "{")
	closeIdx := strings.LastIndex(line, "}")
	if openIdx < 0 || closeIdx < 0 || closeIdx <= openIdx {
		return nil, fmt.Errorf("invalid tags block")
	}

	raw := line[openIdx+1 : closeIdx]
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		tag = trimQuoted(tag)
		if tag != "" {
			out = append(out, tag)
		}
	}

	return out, nil
}

func trimQuoted(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"")
	return value
}
