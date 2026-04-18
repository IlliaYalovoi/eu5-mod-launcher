package mods

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

var (
	errUnterminatedTagsBlock = errors.New("unterminated tags block")
	errInvalidTagsBlock      = errors.New("invalid tags block")
)

const (
	jsonControlByteLimit = 0x20
	decimalBase          = 10
	utf16BOMByteLen      = 4
)

// WorkshopMetadata represents the metadata for a Steam Workshop mod.
type WorkshopMetadata struct {
	Title            string   `json:"title"`
	ShortDescription string   `json:"shortDescription"`
	Tags             []string `json:"tags"`
	FileID           string   `json:"fileId"`
}

// Descriptor holds the metadata fields for a mod.
type Descriptor struct {
	Name             string
	Version          string
	SupportedVersion string
	Description      string
	Tags             []string
}

// ParseDescriptor reads a descriptor.mod file and fills mod metadata fields.
// Unknown keys are silently ignored.
func ParseDescriptor(path string) (Descriptor, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Descriptor{}, fmt.Errorf("read descriptor %q: %w", path, err)
	}

	normalized := normalizeDescriptorBytes(content)

	if strings.EqualFold(filepath.Ext(path), ".json") {
		parsed, parseErr := parseJSONDescriptor(normalized)
		if parseErr != nil {
			return Descriptor{}, fmt.Errorf("parse json descriptor %q: %w", path, parseErr)
		}
		return parsed, nil
	}

	parsed, parseErr := parseTextDescriptor(string(normalized))
	if parseErr != nil {
		return Descriptor{}, fmt.Errorf("parse text descriptor %q: %w", path, parseErr)
	}

	return parsed, nil
}

func parseJSONDescriptor(content []byte) (Descriptor, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(content, &payload); err != nil {
		sanitized := sanitizeBrokenJSONStringLiterals(content)
		if bytes.Equal(sanitized, content) {
			return Descriptor{}, fmt.Errorf("unmarshal descriptor json: %w", err)
		}

		if sanitizeErr := json.Unmarshal(sanitized, &payload); sanitizeErr != nil {
			return Descriptor{}, fmt.Errorf("unmarshal descriptor json: %w", err)
		}
	}

	name := extractJSONString(payload, "name")
	version := extractJSONString(payload, "version")
	supportedVersion := extractJSONString(payload, "supported_version")
	description := extractJSONString(payload, "description")
	if description == "" {
		description = extractAnyJSONString(payload, "shortDescription", "short_description")
	}
	tags := extractJSONStringArray(payload, "tags")

	return Descriptor{
		Name:             name,
		Version:          version,
		SupportedVersion: supportedVersion,
		Description:      description,
		Tags:             tags,
	}, nil
}

func extractAnyJSONString(payload map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		value := extractJSONString(payload, key)
		if value != "" {
			return value
		}
	}
	return ""
}

func extractJSONString(payload map[string]json.RawMessage, key string) string {
	raw, ok := payload[key]
	if !ok {
		return ""
	}
	var out string
	if err := json.Unmarshal(raw, &out); err != nil {
		return ""
	}
	return out
}

func extractJSONStringArray(payload map[string]json.RawMessage, key string) []string {
	raw, ok := payload[key]
	if !ok {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func sanitizeBrokenJSONStringLiterals(content []byte) []byte {
	output := make([]byte, 0, len(content)+32)
	inString := false
	escaped := false
	changed := false

	for _, currentByte := range content {
		if !inString {
			output = append(output, currentByte)
			if currentByte == '"' {
				inString = true
			}
			continue
		}

		if escaped {
			output = append(output, currentByte)
			escaped = false
			continue
		}

		switch currentByte {
		case '\\':
			output = append(output, currentByte)
			escaped = true
		case '"':
			output = append(output, currentByte)
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
			if currentByte < jsonControlByteLimit {
				output = append(output, '\\', 'u', '0', '0', hexDigit(currentByte>>4), hexDigit(currentByte&0x0F))
				changed = true
				continue
			}
			output = append(output, currentByte)
		}
	}

	if !changed {
		return content
	}

	return output
}

func hexDigit(nibble byte) byte {
	if nibble < decimalBase {
		return '0' + nibble
	}
	return 'a' + (nibble - decimalBase)
}

func parseTextDescriptor(content string) (Descriptor, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	parsed := Descriptor{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "tags") {
			parsedTags, parseErr := parseTagsBlock(line, scanner)
			if parseErr != nil {
				return Descriptor{}, fmt.Errorf("parse tags block: %w", parseErr)
			}
			parsed.Tags = parsedTags
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
			parsed.Name = value
		case "version":
			parsed.Version = value
		case "supported_version":
			parsed.SupportedVersion = value
		case "description", "short_description":
			parsed.Description = value
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return Descriptor{}, fmt.Errorf("scan descriptor lines: %w", scanErr)
	}

	return parsed, nil
}

func normalizeDescriptorBytes(content []byte) []byte {
	if bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}) {
		return content[3:]
	}

	hasUTF16BOM := bytes.HasPrefix(content, []byte{0xFF, 0xFE}) ||
		bytes.HasPrefix(content, []byte{0xFE, 0xFF})
	if len(content) >= 2 && hasUTF16BOM {
		return decodeUTF16Descriptor(content)
	}

	return content
}

func decodeUTF16Descriptor(content []byte) []byte {
	if len(content) < utf16BOMByteLen {
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
			return nil, errUnterminatedTagsBlock
		}

		builder.WriteString("\n")
		builder.WriteString(strings.TrimSpace(scanner.Text()))
	}

	line := builder.String()
	openIdx := strings.Index(line, "{")
	closeIdx := strings.LastIndex(line, "}")
	if openIdx < 0 || closeIdx < 0 || closeIdx <= openIdx {
		return nil, errInvalidTagsBlock
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
