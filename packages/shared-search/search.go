package search

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type Client interface {
	Index(ctx context.Context, indexName string, docID string, body interface{}) error
	Search(ctx context.Context, indexName string, query string, companyID string) ([]SearchResult, error)
}

type SearchResult struct {
	TraceID     string `json:"traceId"`
	NodeID      string `json:"nodeId"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	DBQuery     string `json:"dbQuery,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`
}

type clientImpl struct {
	osClient *opensearch.Client
}

func NewClient(addresses []string) (Client, error) {
	osClient, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: addresses,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create opensearch client: %w", err)
	}

	return &clientImpl{osClient: osClient}, nil
}

func (c *clientImpl) Index(ctx context.Context, indexName string, docID string, body interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       strings.NewReader(string(jsonBody)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.osClient)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("indexing error [%s]: %s", res.Status(), string(b))
	}

	return nil
}

func (c *clientImpl) Search(ctx context.Context, indexName string, query string, companyID string) ([]SearchResult, error) {
	// A simple query string search on the whole document, filtered by companyId
	q := fmt.Sprintf(`{
		"query": {
			"bool": {
				"must": [
					{
						"multi_match": {
							"query": "%s",
							"fields": ["*"]
						}
					}
				],
				"filter": [
					{
						"term": {
							"companyId.keyword": "%s"
						}
					}
				]
			}
		},
		"_source": ["traceId", "nodeId", "name", "type", "dbQuery", "metadata", "serviceName"]
	}`, query, companyID)

	req := opensearchapi.SearchRequest{
		Index: []string{indexName},
		Body:  strings.NewReader(q),
	}

	res, err := req.Do(ctx, c.osClient)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error [%s]: %s", res.Status(), string(b))
	}

	var r struct {
		Hits struct {
			Hits []struct {
				Source SearchResult `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var results []SearchResult
	for _, hit := range r.Hits.Hits {
		results = append(results, hit.Source)
	}

	return results, nil
}
