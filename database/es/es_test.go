package es

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var client *ElasticClient

func init() {
	cfg := &Config{Addresses: []string{"http://127.0.0.1:9200"}}
	client = NewElasticsearch(cfg)
}

// go test -v -test.run TestDatabase
func TestDatabase(t *testing.T) {
	res, err := client.client.Info()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("close info response: %v", err)
		}
	}()
	if res.IsError() {
		t.Fatalf("elasticsearch info returned %s", res.Status())
	}
}

// go test -v -test.run TestCreateIndex
func TestQueryAndLegacyAlias(t *testing.T) {
	var queryCalls int
	server := newElasticTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/_search") {
			queryCalls++
			writeJSON(w, http.StatusOK, map[string]any{"hits": map[string]any{"total": map[string]any{"value": 0}}})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"version": map[string]any{"number": "8.0.0"}})
	})
	defer server.Close()

	testClient, err := Open(&Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := testClient.Query("items", map[string]any{"query": map[string]any{"match_all": map[string]any{}}}, &result); err != nil {
		t.Fatal(err)
	}
	if err := testClient.Qurey("items", map[string]any{}, &result); err != nil {
		t.Fatal(err)
	}
	if queryCalls != 2 {
		t.Fatalf("query calls = %d, want 2", queryCalls)
	}
}

// go test -v -test.run TestCreateIndexResponseOwnership
func TestCreateIndexResponseOwnership(t *testing.T) {
	server := newElasticTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"acknowledged": true})
	})
	defer server.Close()
	testClient, err := Open(&Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatal(err)
	}
	res, err := testClient.CreateIndex("items", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("close create index response: %v", err)
		}
	}()
	if _, err := io.ReadAll(res.Body); err != nil {
		t.Fatalf("successful response body was not readable: %v", err)
	}
}

// go test -v -test.run TestElasticsearchStatusError
func TestElasticsearchStatusError(t *testing.T) {
	server := newElasticTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			writeJSON(w, http.StatusOK, map[string]any{"version": map[string]any{"number": "8.0.0"}})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bad request"})
	})
	defer server.Close()
	testClient, err := Open(&Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = testClient.CreateIndex("items", `{}`)
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("CreateIndex error = %v, want status 400", err)
	}
}

// go test -v -test.run TestElasticsearchResponseError
func TestUpdateDocumentByQuery(t *testing.T) {
	paths := make([]string, 0, 2)
	server := newElasticTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			writeJSON(w, http.StatusOK, map[string]any{"version": map[string]any{"number": "8.0.0"}})
			return
		}
		paths = append(paths, r.URL.Path)
		writeJSON(w, http.StatusOK, map[string]any{"updated": 1})
	})
	defer server.Close()
	testClient, err := Open(&Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatal(err)
	}
	if err := testClient.UpdateDocumentByQuery("items", `{}`); err != nil {
		t.Fatal(err)
	}
	if err := testClient.UpdateDocumentByQurey(`{}`); err != nil {
		t.Fatal(err)
	}
	want := []string{"/items/_update_by_query", "/test_index/_update_by_query"}
	if len(paths) != len(want) {
		t.Fatalf("update paths = %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("update path %d = %q, want %q", i, paths[i], want[i])
		}
	}
}

// go test -v -test.run TestElasticsearchResponseError
func newElasticTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		handler(w, r)
	}))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
