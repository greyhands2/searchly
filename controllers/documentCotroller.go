package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	aProduct "github.com/greyhands2/searchly/model"
)

var validate = validator.New()

type Collation struct {
	Query string `json:"query"`
}

type Option struct {
	Text   string `json:"text"`
	Freq   int    `json:"freq"`
	Score  int    `json:"score"`
	Length int    `json:length`
}

type Suggestion struct {
	Length    int       `json:"length"`
	Offset    int       `json:"offset"`
	Text      string    `json:"text"`
	Options   []Option  `json:"options"`
	Collation Collation `json:"collation"`
}

type SuggesterResponse struct {
	Suggest map[string][]Suggestion `json:"suggest"`
}

func Search(req_res *fiber.Ctx) error {
	var err error
	ctx := req_res.Locals("ctx").(context.Context)
	es := ctx.Value("esConnection").(*elasticsearch.Client)
	query := req_res.Params("query")

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"product_name": query,
						},
					},
					{
						"match": map[string]interface{}{
							"category": query,
						},
					},
				},
			},
		},

		"suggest": map[string]interface{}{

			"product_name_suggestion": map[string]interface{}{
				"text": query,
				"term": map[string]interface{}{
					"field": "product_name", // Replace with the field you want to suggest on
				},
				"phrase": map[string]interface{}{
					"field":     "product_name.phrase", // Replace with the field you want to suggest on
					"size":      5,
					"gram_size": 1,
					"direct_generator": []map[string]interface{}{
						{
							"field":           "product_name.phrase",
							"suggest_mode":    "always",
							"min_word_length": 1,
						},
					},
					"highlight": map[string]interface{}{
						"pre_tag":  "<em>",
						"post_tag": "</em>",
					},
				},
			},
			"category_suggestion": map[string]interface{}{
				"text": query,
				"term": map[string]interface{}{
					"field": "category",
				},
				"phrase": map[string]interface{}{
					"field":     "category.phrase",
					"size":      5,
					"gram_size": 1,
					"direct_generator": []map[string]interface{}{
						{
							"field":           "category.phrase",
							"suggest_mode":    "always",
							"min_word_length": 1,
						},
					},
					"highlight": map[string]interface{}{
						"pre_tag":  "<em>",
						"post_tag": "</em>",
					},
				},
			},
		},
	}

	searchQueryJSON, err := json.Marshal(searchQuery)
	if err != nil {
		fmt.Println(err)
		return req_res.Status(500).SendString("Something Went Wrong ")
	}

	req := esapi.SearchRequest{
		Index: []string{"products"},
		Body:  bytes.NewBuffer(searchQueryJSON),
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		fmt.Println(err)
		return req_res.Status(500).SendString("Something Went Wrong ")
	}

	defer res.Body.Close()
	if res.IsError() {
		fmt.Println(res.Status())
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(body))
		return req_res.Status(500).SendString("Something Went Wrong ")
	}

	var searchResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return req_res.Status(500).SendString("Something Went Wrong ")
	}

	log.Println("Search Results:", searchResponse)
	return req_res.Status(200).JSON(searchResponse)

}

func InsertDocument(req_res *fiber.Ctx) error {
	var err error
	var product aProduct.Product

	ctx := req_res.Locals("ctx").(context.Context)
	es := ctx.Value("esConnection").(*elasticsearch.Client)
	// collect request body into user struct and handle error

	if err = req_res.BodyParser(&product); err != nil {
		return req_res.Status(400).SendString(err.Error())
	}

	//now let's validate the data using the struct
	structValidationError := validate.Struct(product)

	if structValidationError != nil {
		return req_res.Status(400).SendString("Data Validation Error")
	}

	product.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	product.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	//convert struct back to json for elasticsearch indexing
	productJSON, err := json.Marshal(product)
	if err != nil {
		return req_res.Status(500).SendString("Something Went Wrong")
	}

	//generate randon string for unique identifier
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 10)
	rand.Read(b)
	var randomString string = fmt.Sprintf("%x", b)[:10]
	fmt.Println(productJSON)
	//now lets insert the product
	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: randomString,
		Body:       bytes.NewReader(productJSON),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return req_res.Status(500).SendString("Something Went Wrong")
	}

	defer res.Body.Close()

	if res.IsError() {
		return req_res.Status(500).SendString("Something Went Wrong")
	}

	return req_res.Status(200).SendString("Product Successfully Stored")

}
