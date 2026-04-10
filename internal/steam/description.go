package steam

import (
	"regexp"
	"sort"
	"strings"
)

var descriptionImageTag = regexp.MustCompile(`(?is)\[img\]\s*(.*?)\s*\[/img\]`)

const (
	imgSubmatchCount    = 2
	imgURLSubmatchIndex = 1
)

func ExtractDescriptionImageURLs(description string) []string {
	matches := descriptionImageTag.FindAllStringSubmatch(description, -1)
	if len(matches) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(matches))
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < imgSubmatchCount {
			continue
		}
		url := strings.TrimSpace(match[imgURLSubmatchIndex])
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		result = append(result, url)
	}

	sort.Strings(result)
	return result
}

func ReplaceDescriptionImageURLs(description string, replacements map[string]string) string {
	if len(replacements) == 0 {
		return description
	}

	return descriptionImageTag.ReplaceAllStringFunc(description, func(match string) string {
		subMatch := descriptionImageTag.FindStringSubmatch(match)
		if len(subMatch) < imgSubmatchCount {
			return match
		}
		current := strings.TrimSpace(subMatch[imgURLSubmatchIndex])
		replacement, ok := replacements[current]
		if !ok || strings.TrimSpace(replacement) == "" {
			return match
		}
		return "[img]" + replacement + "[/img]"
	})
}
