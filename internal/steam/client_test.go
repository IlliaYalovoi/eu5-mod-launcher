package steam_test

import (
	"eu5-mod-launcher/internal/steam"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUserAgent = "eu5-mod-launcher-test"

func TestClientFetchWorkshopMetadata_Success(t *testing.T) {
	t.Parallel()

	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, testUserAgent, r.UserAgent())

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "itemcount=2")
		assert.Contains(t, string(body), "publishedfileids%5B0%5D=123")
		assert.Contains(t, string(body), "publishedfileids%5B1%5D=456")

		_, writeErr := w.Write([]byte(`{"response":{"publishedfiledetails":[` +
			`{"publishedfileid":"123","result":1,"title":"A","description":"Desc A","preview_url":"https://img/a"},` +
			`{"publishedfileid":"456","result":1,"title":"B","description":"Desc B","preview_url":"https://img/b"}` +
			`]}}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	client := steam.NewClientWithOptions(&http.Client{Timeout: 5 * time.Second}, server.URL, testUserAgent, 0, 0)
	items, err := client.FetchWorkshopMetadata([]string{"123", "456", "123"})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
	assert.Equal(t, "A", items["123"].Title)
	assert.Equal(t, "Desc B", items["456"].Description)
	assert.Equal(t, "https://img/b", items["456"].PreviewURL)
}

func TestClientFetchWorkshopMetadata_RetriesRetryableStatus(t *testing.T) {
	t.Parallel()

	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt := atomic.AddInt32(&callCount, 1)
		if attempt == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, writeErr := w.Write([]byte("temporary"))
			require.NoError(t, writeErr)
			return
		}
		_, writeErr := w.Write([]byte(`{"response":{"publishedfiledetails":[{"publishedfileid":"123","result":1,"title":"A"}]}}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	client := steam.NewClientWithOptions(
		&http.Client{Timeout: 5 * time.Second},
		server.URL,
		testUserAgent,
		1,
		1*time.Millisecond,
	)

	items, err := client.FetchWorkshopMetadata([]string{"123"})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount))
}

func TestClientFetchWorkshopMetadata_InvalidID(t *testing.T) {
	t.Parallel()

	client := steam.NewClientWithOptions(&http.Client{Timeout: 5 * time.Second}, "http://localhost", testUserAgent, 0, 0)
	_, err := client.FetchWorkshopMetadata([]string{"abc"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workshop item id")
}

func TestClientFetchWorkshopMetadata_NonRetryableStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := w.Write([]byte("bad request"))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	client := steam.NewClientWithOptions(
		&http.Client{Timeout: 5 * time.Second},
		server.URL,
		testUserAgent,
		2,
		1*time.Millisecond,
	)

	_, err := client.FetchWorkshopMetadata([]string{"123"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 400")
}

func TestClientFetchWorkshopMetadata_ParseValidation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, writeErr := w.Write([]byte(`{"response":{}}`))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	client := steam.NewClientWithOptions(&http.Client{Timeout: 5 * time.Second}, server.URL, testUserAgent, 0, 0)
	_, err := client.FetchWorkshopMetadata([]string{"123"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "publishedfiledetails")
}
