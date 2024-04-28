// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package rpc

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/google/uuid"

	//zinc "github.com/zinclabs/sdk-go-zincsearch"
	"io/ioutil"
	//"github.com/elastic/go-elasticsearch/v8/esapi"
	builtinlog "log"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/rs/zerolog/log"
)

var ErrQuery = errors.New("query err")

type RssQueryResult struct {
	Index       string   `json:"index"`
	Name        string   `json:"name"`
	DocId       string   `json:"docId"`
	Created     int64    `json:"created"`
	Content     string   `json:"content"`
	Meta        string   `json:"meta"`
	HightLights []string `json:"highlight"`
}

func InitES(url string, username string, password string) (es *elasticsearch.TypedClient, err error) {
	builtinlog.SetFlags(0)

	// This example demonstrates how to configure the client's Transport.
	//
	// NOTE: These values are for illustrative purposes only, and not suitable
	//       for any production use. The default transport is sufficient.
	//
	if url == "" {
		url = "http://localhost:4080"
	}
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "User#123"
	}

	cfg := elasticsearch.Config{
		Addresses: []string{url + "/es/"},
		Username:  username,
		Password:  password,
		//Transport: &http.Transport{
		//	MaxIdleConnsPerHost:   10,
		//	ResponseHeaderTimeout: time.Millisecond,
		//	DialContext:           (&net.Dialer{Timeout: time.Nanosecond}).DialContext,
		//	TLSClientConfig: &tls.Config{
		//		MinVersion: tls.VersionTLS12,
		//		// ...
		//	},
		//},
	}

	es, err = elasticsearch.NewTypedClient(cfg)
	if err != nil {
		builtinlog.Printf("Error creating the client: %s", err)
	} else {
		builtinlog.Println(es.Info())
		// => dial tcp: i/o timeout
	}
	return
}

func (s *Service) EsSetupIndex() error {

	//"content": {
	//	"type": "text",
	//		"index": true,
	//		"store": true,
	//		"sortable": false,
	//		"aggregatable": false,
	//		"highlightable": true
	//},
	//"md5": {
	//	"type": "text",
	//		"analyzer": "keyword",
	//		"index": true,
	//		"store": false,
	//		"sortable": false,
	//		"aggregatable": false,
	//		"highlightable": false
	//},
	//"where": {
	//	"type": "text",
	//		"analyzer": "keyword",
	//		"index": true,
	//		"store": false,
	//		"sortable": false,
	//		"aggregatable": false,
	//		"highlightable": false
	//}

	expectIndexList := []string{RssIndex}

	// 创建映射定义
	mapping := `
	{
	    "content": {
          "type": "text",
		  "index": true,
          "store": true,
          "sortable": false,
          "aggregatable": false,
          "highlightable": true
	    },
	    "md5": {
          "type": "text",
		  "analyzer": "keyword",
          "index": true,
          "store": false,
          "sortable": false,
          "aggregatable": false,
          "highlightable": false
	    },
	    "where": {
          "type": "text",
		  "analyzer": "keyword",
	      "index": true,
          "store": false,
          "sortable": false,
          "aggregatable": false,
          "highlightable": false
	    }
	}
	`
	var propertiesMap map[string]types.Property
	//{
	//	ContentFieldName: types.TextProperty{},
	//	"md5":            types.KeywordProperty{},
	//	"where":          types.KeywordProperty{},
	//}
	fmt.Println(propertiesMap)
	err := json.Unmarshal([]byte(mapping), &propertiesMap)
	if err != nil {
		return err
	}
	fmt.Println(propertiesMap)

	var NewIndexTypedMapping *types.TypeMapping = &types.TypeMapping{
		//Properties: map[string]types.Property{
		//	ContentFieldName: types.NewTextProperty(),
		//	"md5":            types.NewKeywordProperty(),
		//	"where":          types.NewKeywordProperty(),
		//},
		Properties: propertiesMap,
	}

	for _, indexName := range expectIndexList {
		//检查索引是否存在
		exists, err := s.esClient.Indices.Exists(indexName).IsSuccess(s.context)
		if err != nil {
			log.Fatal().Msgf("check index exist failed", err.Error())
		}

		//索引不存在则创建索引
		//索引不存在时查询会报错，但索引不存在的时候可以直接插入
		if !exists {
			log.Info().Msgf("index %s is not exist, to create", indexName)
			cr, err := s.esClient.Indices.Create(indexName).Request(&create.Request{
				Mappings: NewIndexTypedMapping,
			}).Do(s.context)
			if err != nil {
				log.Fatal().Msgf("create index failed", err.Error())
			}
			log.Info().Msgf("index create", cr.Index)
		}

		//// 设置索引的映射
		//var propertiesMap map[string]types.Property
		//err = json.Unmarshal([]byte(mapping), &propertiesMap)
		//if err != nil {
		//	return err
		//}
		//fmt.Println(propertiesMap)
		//req := s.esClient.Indices.PutMapping(indexName).Properties(propertiesMap)
		//res, err := req.Perform(s.context)
		//if err != nil {
		//	log.Info().Msgf("Error getting response: %s", err)
		//}
		//defer res.Body.Close()
		//
		//fmt.Println("Mapping created successfully.")
	}
	return nil
}

func (s *Service) EsDelete(docId string, index string) ([]byte, error) {
	res, err := s.esClient.Delete(index, docId).Perform(s.context)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, ErrQuery
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *Service) EsInput(index string, document map[string]interface{}) ([]byte, error) {
	docId := uuid.NewString()
	fmt.Println(docId)
	//fmt.Println(esutil.NewJSONReader(document))
	res, err := s.esClient.Index(index).Raw(esutil.NewJSONReader(document)).Id(docId).Perform(s.context)
	if err != nil {
		log.Fatal().Msgf("Error indexing document: %s", err)
	}
	defer res.Body.Close()

	fmt.Println("Document indexed successfully. Document ID:", docId)
	return []byte(docId), nil
}

func (s *Service) EsRawQuery(indexName, term string, size int) (*search.Response, error) {
	res, err := s.esClient.Search().
		Index(indexName).
		Request(&search.Request{
			Query: &types.Query{
				Bool: &types.BoolQuery{
					Should: []types.Query{
						{
							Match: map[string]types.MatchQuery{
								"content": {Query: term},
							},
						},
						{
							Match: map[string]types.MatchQuery{
								"format_name": {Query: term},
							},
						},
						{
							Match: map[string]types.MatchQuery{
								"name": {Query: term},
							},
						},
					},
				},
			},
			Size: &size,
			Highlight: &types.Highlight{
				Fields: map[string]types.HighlightField{
					"content": {},
				},
			},
		}).Do(s.context)
	if err != nil {
		return nil, fmt.Errorf("error when calling `SearchApi.Search``: %v", err)
	}
	for index, value := range res.Hits.Hits {
		fmt.Println(index, res.Fields, res.UnmarshalJSON(value.Source_))
	}
	fmt.Println(res.Hits.Total)
	return res, nil
}

func EsGetRssQueryResult(resp *search.Response) ([]RssQueryResult, error) {
	resultList := make([]RssQueryResult, 0)
	for _, hit := range resp.Hits.Hits {
		result := RssQueryResult{
			Index:       RssIndex,
			DocId:       hit.Id_,
			HightLights: make([]string, 0),
		}
		var data map[string]interface{}
		err := json.Unmarshal(hit.Source_, &data)
		if err != nil {
			fmt.Println("解析 Source 字段时出错:", err)
			continue
		}
		if name, ok := data["name"].(string); ok {
			result.Name = name
		}
		if created, ok := data["created"].(float64); ok {
			result.Created = int64(created)
		}
		if content, ok := data["content"].(string); ok {
			result.Content = content
		}
		if meta, ok := data["meta"].(string); ok {
			result.Meta = meta
		}

		for _, highlightRes := range hit.Highlight {
			for _, h := range highlightRes {
				result.HightLights = append(result.HightLights, h)
			}
		}
		resultList = append(resultList, result)
	}
	return resultList, nil
}
