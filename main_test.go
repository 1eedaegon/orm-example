package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/1eedaegon/orm-example/ent/enttest"
	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	// 테스트용 memory rdbms, Sqlite3을 생성한다.
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	err := seed(context.Background(), client)
	require.NoError(t, err)

	srv := NewServer(client)
	r := NewRouter(srv)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test request HTTP GET to "/"
	resp, err := ts.Client().Get(ts.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Test response body has "hello world"
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Hello world!")
}
