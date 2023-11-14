package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/greyhands2/searchly/vipr"
)

var userName string = vipr.ViperEnvVariable("ELASTIC_USER")
var password string = vipr.ViperEnvVariable("ELASTIC_PASSWORD")

var port string = vipr.ViperEnvVariable("ES_PORT")
var esHost string = vipr.ViperEnvVariable("ES_HOST")
var address string = fmt.Sprintf("%s:%s", esHost, port)

func ElasticSearchConnect() (*elasticsearch.Client, error) {
	cert, _ := ioutil.ReadFile("certificates/ca/ca.crt")

	// Create the Elasticsearch client
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
		Username:  userName,
		Password:  password,
		CACert:    cert,
	})

	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
		return nil, err
	}

	// Ping the Elasticsearch server
	test, err := es.Ping()
	if err != nil {
		log.Fatalf("Error pinging Elasticsearch: %s", err)
		return nil, err
	}

	defer test.Body.Close()

	if test.IsError() {
		log.Fatalf("Error: %s", test.String())
		return nil, fmt.Errorf("Elasticsearch ping was unsuccessful")
	}

	log.Println("Connected to Elasticsearch")

	// Check if the index 'products' exists
	indexExists := esapi.IndicesExistsRequest{
		Index: []string{"products"},
	}
	existsResponse, err := indexExists.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error checking if index exists: %s", err)
		return nil, err
	}

	if existsResponse.StatusCode == 404 {
		// The index does not exist; create it
		mapping := `{
			"mappings": {
			  "properties": {
				"price":{
					"type": "float"
				},
				"product_name": {
				  "type": "text",
				  "fields": {
					"completion": {
						"type": "completion",
						  "analyzer": "simple",
						  "search_analyzer": "standard",
						  "preserve_separators": true,
						  "preserve_position_increments": true
					},
					"phrase": {
						"type": "text",
						"analyzer": "trigram",
						"term_vector": "yes"
					  }

				}

				},
				"category": {
				"type": "text",
				"fields": {
				  "completion": {
					  "type": "completion",
					  "analyzer": "simple",
					  "search_analyzer": "standard",
					  "preserve_separators": true,
					  "preserve_position_increments": true
				  },
				  "phrase": {
					"type": "text",
					"analyzer": "trigram",
					"term_vector": "yes"
				  }
				 
				}
			}

		  }

		},
		"settings": {
			"index": {
			  "number_of_shards": 1,
			  "analysis": {
				"analyzer": {
				  "trigram": {
					"type": "custom",
					"tokenizer": "standard",
					"filter": ["lowercase","shingle"]
				  },
				  "reverse": {
					"type": "custom",
					"tokenizer": "standard",
					"filter": ["lowercase","reverse"]
				  }
				},
				"filter": {
				  "shingle": {
					"type": "shingle",
					"min_shingle_size": 2,
					"max_shingle_size": 3
				  }
				}
			  }
			}
		  }

		}
		`

		createIndexRequest := esapi.IndicesCreateRequest{
			Index: "products",
			Body:  strings.NewReader(mapping),
		}

		createIndexResponse, err := createIndexRequest.Do(context.Background(), es)
		if err != nil {
			log.Fatalf("Error creating index: %s", err)
			return nil, err
		}

		defer createIndexResponse.Body.Close()
	}

	return es, nil

}
