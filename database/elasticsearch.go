package database

import (
	"bbsDemo/config"
	"bbsDemo/logger"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"
)

var (
	ESClient *elasticsearch.Client
)

type ESDocument struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // post or comment
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	PostID    int64     `json:"post_id,omitempty"`
	CommentID uint      `json:"comment_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func InitElasticsearch(cfg config.ElasticsearchConfig) error {
	esConfig := elasticsearch.Config{
		Addresses: cfg.Hosts,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		logger.Error("Failed to create Elasticsearch client", zap.Error(err))
		return err
	}

	ESClient = client

	// 检查连接
	resp, err := ESClient.Info()
	if err != nil {
		logger.Error("Failed to check Elasticsearch connection", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("Elasticsearch info error: %s", resp.Status())
	}

	// 创建索引
	if err := createIndex(cfg.Index); err != nil {
		logger.Error("Failed to create Elasticsearch index", zap.Error(err))
		return err
	}

	logger.Info("Elasticsearch connected successfully",
		zap.Strings("hosts", cfg.Hosts),
		zap.String("index", cfg.Index))

	return nil
}

func createIndex(index string) error {
	settings := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"default": map[string]interface{}{
						"type": "ik_max_word",
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"type": map[string]interface{}{
					"type": "keyword",
				},
				"title": map[string]interface{}{
					"type":     "text",
					"analyzer": "ik_max_word",
				},
				"content": map[string]interface{}{
					"type":     "text",
					"analyzer": "ik_max_word",
				},
				"user_id": map[string]interface{}{
					"type": "long",
				},
				"post_id": map[string]interface{}{
					"type": "long",
				},
				"comment_id": map[string]interface{}{
					"type": "long",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader(data),
	}

	resp, err := req.Do(context.Background(), ESClient)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		// 如果索引已存在，忽略错误
		if resp.StatusCode == 400 {
			var errorResp map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&errorResp) == nil {
				if errorResp["error"].(map[string]interface{})["type"] == "resource_already_exists_exception" {
					return nil
				}
			}
		}
		return fmt.Errorf("failed to create index: %s", resp.Status())
	}

	return nil
}

func IndexDocument(index string, docID string, doc ESDocument) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	resp, err := req.Do(context.Background(), ESClient)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("failed to index document: %s", resp.Status())
	}

	return nil
}

func SearchDocuments(index string, query string, size int) ([]ESDocument, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title", "content"},
				"type":   "best_fields",
			},
		},
		"size": size,
	}

	data, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(data),
	}

	resp, err := req.Do(context.Background(), ESClient)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, fmt.Errorf("search failed: %s", resp.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var documents []ESDocument
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		doc := ESDocument{
			ID:      hitMap["_id"].(string),
			Type:    source["type"].(string),
			Content: source["content"].(string),
			UserID:  int64(source["user_id"].(float64)),
		}

		if title, ok := source["title"]; ok {
			doc.Title = title.(string)
		}

		if postID, ok := source["post_id"]; ok {
			doc.PostID = int64(postID.(float64))
		}

		if commentID, ok := source["comment_id"]; ok {
			doc.CommentID = uint(commentID.(float64))
		}

		if createdAt, ok := source["created_at"]; ok {
			createdAtStr := createdAt.(string)
			createdAtTime, _ := time.Parse(time.RFC3339, createdAtStr)
			doc.CreatedAt = createdAtTime
		}

		if updatedAt, ok := source["updated_at"]; ok {
			updatedAtStr := updatedAt.(string)
			updatedAtTime, _ := time.Parse(time.RFC3339, updatedAtStr)
			doc.UpdatedAt = updatedAtTime
		}

		documents = append(documents, doc)
	}

	return documents, nil
}

func CloseElasticsearch() error {
	if ESClient != nil {
		// Elasticsearch client doesn't have a Close method
		// Just set to nil
		ESClient = nil
		logger.Info("Elasticsearch client closed")
	}
	return nil
}
