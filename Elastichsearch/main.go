package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var es *elasticsearch.Client

func init() {
	// 初始化 Elasticsearch 客户端
	var err error
	es, err = elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
}

func createIndex() {
	// 创建索引
	req := esapi.IndicesCreateRequest{
		Index: "test-index",
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error creating index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error creating index: %s", res.Status())
	}

	fmt.Println("Index created successfully.")
}

func indexDocument() {
	// 索引一个文档
	document := map[string]interface{}{
		"user":    "John",
		"message": "Hello, Elasticsearch!",
	}

	// 将文档转为 JSON 格式
	body, err := json.Marshal(document)
	if err != nil {
		log.Fatalf("Error marshaling document: %s", err)
	}

	req := esapi.IndexRequest{
		Index:      "test-index",
		DocumentID: "1",
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	// 执行请求
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error indexing document: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error indexing document: %s", res.Status())
	}

	fmt.Println("Document indexed successfully.")
}

func searchDocument() {
	// 搜索文档
	req := esapi.SearchRequest{
		Index: []string{"test-index"},
		Body:  bytes.NewReader([]byte(`{"query": {"match": {"message": "Hello"}}}`)),
	}

	// 执行搜索请求
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error executing search: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error executing search: %s", res.Status())
	}

	// 打印响应
	fmt.Println("Search Results:")
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	// 输出查询结果
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		fmt.Printf("ID: %v, Source: %v\n", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}
}

func deleteDocument() {
	// 删除文档
	req := esapi.DeleteRequest{
		Index:      "test-index",
		DocumentID: "1",
	}

	// 执行删除请求
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error deleting document: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error deleting document: %s", res.Status())
	}

	fmt.Println("Document deleted successfully.")
}

func main() {
	// 创建索引
	createIndex()

	// 索引一个文档
	indexDocument()

	// 查询文档
	searchDocument()

	// 删除文档
	deleteDocument()
}

// 注意ES可以localhost：9200访问但使用不了，考虑是不是防火墙的问题
