package opensearch

import (
	"context"
	"fmt"
	"os"

	"github.com/opensearch-project/opensearch-go/v3"
	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"
)

// ConnectOpenSearch initializes and returns an OpenSearch v3 client.
func ConnectOpenSearch() (*opensearchapi.Client, error) {
	host := os.Getenv("OPENSEARCH_HOST")
	port := os.Getenv("OPENSEARCH_PORT")
	user := os.Getenv("OPENSEARCH_USER")
	pass := os.Getenv("OPENSEARCH_PASSWORD")

	if host == "" || port == "" {
		return nil, fmt.Errorf("OPENSEARCH_HOST or OPENSEARCH_PORT not set")
	}

	// For online hosted OpenSearch, we usually use https.
	// Construct address: https://host:port
	address := fmt.Sprintf("https://%s:%s", host, port)

	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Addresses: []string{address},
			Username:  user,
			Password:  pass,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	// Ping the cluster to verify connectivity
	_, err = client.Info(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping OpenSearch: %w", err)
	}

	return client, nil
}
