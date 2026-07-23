package es

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Config configures the official Elasticsearch client.
type Config struct {
	Addresses []string
	Username  string
	Password  string
	APIKey    string
}

// ElasticClient wraps the official Elasticsearch client with small helpers.
type ElasticClient struct {
	conf   *Config
	client *elasticsearch.Client
}

// NewElasticsearch creates an Elasticsearch client and panics on failure.
func NewElasticsearch(c *Config) *ElasticClient {
	client, err := Open(c)
	if err != nil {
		panic(fmt.Errorf("open elasticsearch: %w", err))
	}
	return client
}

// Open creates an Elasticsearch client and verifies the cluster is reachable.
func Open(c *Config) (*ElasticClient, error) {
	if c == nil {
		return nil, errors.New("es: nil config")
	}
	if len(c.Addresses) == 0 {
		return nil, errors.New("es: no addresses configured")
	}
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: c.Addresses,
		Username:  c.Username,
		Password:  c.Password,
		APIKey:    c.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("get elasticsearch info: %w", err)
	}
	if res.IsError() {
		return nil, responseError("get elasticsearch info", res)
	}
	if err := res.Body.Close(); err != nil {
		return nil, fmt.Errorf("close elasticsearch info response: %w", err)
	}
	return &ElasticClient{conf: c, client: es}, nil
}

// CreateIndex creates an index. The caller owns the successful response body.
func (e *ElasticClient) CreateIndex(index, body string) (*esapi.Response, error) {
	res, err := (&esapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(body),
	}).Do(context.Background(), e.client)
	if err != nil {
		return nil, fmt.Errorf("create index %q: %w", index, err)
	}
	if res.IsError() {
		return nil, responseError("create index "+index, res)
	}
	return res, nil
}

// DeleteIndex deletes indexes. The caller owns the successful response body.
func (e *ElasticClient) DeleteIndex(index []string) (*esapi.Response, error) {
	res, err := (&esapi.IndicesDeleteRequest{Index: index}).Do(context.Background(), e.client)
	if err != nil {
		return nil, fmt.Errorf("delete indexes: %w", err)
	}
	if res.IsError() {
		return nil, responseError("delete indexes", res)
	}
	return res, nil
}

func (e *ElasticClient) CreateDocument(index, docID, body string) (err error) {
	res, err := (&esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(body),
		Refresh:    "true",
	}).Do(context.Background(), e.client)
	if err != nil {
		return fmt.Errorf("create document %q: %w", docID, err)
	}
	if res.IsError() {
		return responseError("create document "+docID, res)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close create document response: %w", closeErr)
		}
	}()
	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode create document response: %w", err)
	}
	return nil
}

func (e *ElasticClient) UpdateDocumentByID(index, docID, body string) error {
	res, err := (&esapi.UpdateRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(body),
	}).Do(context.Background(), e.client)
	if err != nil {
		return fmt.Errorf("update document %q: %w", docID, err)
	}
	if res.IsError() {
		return responseError("update document "+docID, res)
	}
	return closeResponse("update document", res)
}

// UpdateDocumentByQuery updates documents matching the supplied query.
func (e *ElasticClient) UpdateDocumentByQuery(index, body string) error {
	res, err := (&esapi.UpdateByQueryRequest{
		Index: []string{index},
		Body:  strings.NewReader(body),
	}).Do(context.Background(), e.client)
	if err != nil {
		return fmt.Errorf("update documents in %q: %w", index, err)
	}
	if res.IsError() {
		return responseError("update documents in "+index, res)
	}
	return closeResponse("update documents", res)
}

// UpdateDocumentByQurey is kept for backward compatibility.
// Deprecated: use UpdateDocumentByQuery.
func (e *ElasticClient) UpdateDocumentByQurey(body string) error {
	return e.UpdateDocumentByQuery("test_index", body)
}

// Query searches an index and decodes the response into result.
func (e *ElasticClient) Query(index string, query map[string]any, result any) (err error) {
	if result == nil {
		return errors.New("es: nil query result")
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return fmt.Errorf("encode query: %w", err)
	}
	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(index),
		e.client.Search.WithBody(&buf),
		e.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return fmt.Errorf("query index %q: %w", index, err)
	}
	if res.IsError() {
		return responseError("query index "+index, res)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close query response: %w", closeErr)
		}
	}()
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return fmt.Errorf("decode query response: %w", err)
	}
	return nil
}

// Qurey is kept for backward compatibility. Deprecated: use Query.
func (e *ElasticClient) Qurey(index string, query map[string]any, result any) error {
	return e.Query(index, query, result)
}

func responseError(operation string, res *esapi.Response) error {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		if closeErr := res.Body.Close(); closeErr != nil {
			return fmt.Errorf("%s: elasticsearch status %s (read body: %v, close body: %v)", operation, res.Status(), err, closeErr)
		}
		return fmt.Errorf("%s: elasticsearch status %s (read body: %v)", operation, res.Status(), err)
	}
	if closeErr := res.Body.Close(); closeErr != nil {
		return fmt.Errorf("%s: elasticsearch status %s: %s (close body: %v)", operation, res.Status(), strings.TrimSpace(string(body)), closeErr)
	}
	return fmt.Errorf("%s: elasticsearch status %s: %s", operation, res.Status(), strings.TrimSpace(string(body)))
}

func closeResponse(operation string, res *esapi.Response) error {
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		_ = res.Body.Close()
		return fmt.Errorf("read %s response: %w", operation, err)
	}
	if err := res.Body.Close(); err != nil {
		return fmt.Errorf("close %s response: %w", operation, err)
	}
	return nil
}
