package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	server200 := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	t.Run("status 200", func(t *testing.T) {
		t.Parallel()
		code, rtt, err := status(server200.URL)
		require.NoError(t, err)
		require.Equal(t, code, http.StatusOK)
		t.Logf("%d - %s (%dms)", code, server200.URL, rtt)
	})

	server404 := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
	t.Run("status 404", func(t *testing.T) {
		t.Parallel()
		code, rtt, err := status(server404.URL)
		require.NoError(t, err)
		require.Equal(t, code, http.StatusNotFound)
		t.Logf("%d - %s (%dms)", code, server404.URL, rtt)
	})

	server501 := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
		}))
	t.Run("status 501", func(t *testing.T) {
		t.Parallel()
		code, rtt, err := status(server501.URL)
		require.NoError(t, err)
		require.Equal(t, code, http.StatusNotImplemented)
		t.Logf("%d - %s (%dms)", code, server501.URL, rtt)
	})
}
