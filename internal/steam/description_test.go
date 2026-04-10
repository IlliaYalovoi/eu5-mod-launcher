package steam_test

import (
	"eu5-mod-launcher/internal/steam"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDescriptionImageURLs(t *testing.T) {
	t.Parallel()

	description := strings.Join([]string{
		"[img]https://example.com/a.png[/img]",
		"text",
		"[img] https://example.com/b.png [/img]",
		"[img]https://example.com/a.png[/img]",
	}, "\n")

	urls := steam.ExtractDescriptionImageURLs(description)
	assert.Equal(
		t,
		[]string{"https://example.com/a.png", "https://example.com/b.png"},
		urls,
	)
}

func TestReplaceDescriptionImageURLs(t *testing.T) {
	t.Parallel()

	description := "[img]https://example.com/a.png[/img]\n[img]https://example.com/b.png[/img]"
	replaced := steam.ReplaceDescriptionImageURLs(description, map[string]string{
		"https://example.com/a.png": "C:\\cache\\a.png",
	})

	assert.Contains(t, replaced, "[img]C:\\cache\\a.png[/img]")
	assert.Contains(t, replaced, "[img]https://example.com/b.png[/img]")
}
