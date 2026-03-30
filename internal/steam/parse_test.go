package steam_test

import (
	"eu5-mod-launcher/internal/steam"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkshopResponseMappingAndFiltering(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, writeErr := w.Write([]byte(`{"response":{"publishedfiledetails":[` +
			`{"publishedfileid":"111","result":1,"title":"Title 1","description":"Desc 1","preview_url":"https://img/1"},` +
			`{"publishedfileid":"222","result":1,"title":"Title 2","description":"Desc 2","preview_url":"https://img/2"},` +
			`{"publishedfileid":"333","result":9,"title":"Ignored"}` +
			`]}}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	client := steam.NewClientWithOptions(
		&http.Client{Timeout: 5 * time.Second},
		server.URL,
		"eu5-mod-launcher-test",
		0,
		0,
	)

	items, err := client.FetchWorkshopMetadata([]string{"111", "333"})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Title 1", items["111"].Title)
	assert.Equal(t, "https://img/1", items["111"].PreviewURL)
}

func TestWorkshopResponseInvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, writeErr := w.Write([]byte(`{"response":`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	client := steam.NewClientWithOptions(
		&http.Client{Timeout: 5 * time.Second},
		server.URL,
		"eu5-mod-launcher-test",
		0,
		0,
	)

	_, err := client.FetchWorkshopMetadata([]string{"111"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode steam response json")
}
