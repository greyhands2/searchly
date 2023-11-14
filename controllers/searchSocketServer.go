package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gofiber/contrib/websocket"
)

type SuggestionResponse struct {
	Suggest map[string][]struct {
		Text    string
		Offset  int
		Length  int
		Options []struct {
			Text  string
			Score float64
		}
	}
}

func SearchSocketServer(c *websocket.Conn) {

	ctx := c.Locals("ctx").(context.Context)
	es := ctx.Value("esConnection").(*elasticsearch.Client)
	var (
		msg []byte
		err error
	)
	for {

		if _, msg, err = c.ReadMessage(); err != nil {

			log.Printf("Error receiving message: %v", err)
			break
		}

		log.Printf("recv: %s", msg)

		msg := string(msg)

		// Handle the "typing" event, perform the search, and send the response
		// Access your Elasticsearch client
		query := msg
		searchQuery := map[string]interface{}{
			"suggest": map[string]interface{}{

				"product_name_suggestion": map[string]interface{}{
					"prefix": query, // Replace with the user's query
					"completion": map[string]interface{}{
						"field": "product_name.completion",
						"size":  5, // Adjust the size as needed
						"fuzzy": map[string]interface{}{
							"fuzziness": "AUTO",
						},
					},
				},
				"category_suggestion": map[string]interface{}{
					"prefix": query, // Replace with the user's query
					"completion": map[string]interface{}{
						"field": "category.completion",
						"size":  5, // Adjust the size as needed
						"fuzzy": map[string]interface{}{
							"fuzziness": "AUTO",
						},
					},
				},
			},
		}

		searchQueryJSON, err := json.Marshal(searchQuery)
		if err != nil {
			log.Println(err)
			return
		}

		req := esapi.SearchRequest{
			Index: []string{"products"},
			Body:  bytes.NewBuffer(searchQueryJSON),
		}

		res, err := req.Do(context.Background(), es)
		if err != nil {
			log.Println(err)
			return
		}

		defer res.Body.Close()
		if res.IsError() {
			log.Println(res.Status())
			body, _ := ioutil.ReadAll(res.Body)
			log.Println(string(body))
			return
		}

		var resultsMap map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&resultsMap); err != nil {
			log.Println(err)
			return
		}

		// Send the search results to the client as a JSON string

		if err = c.WriteJSON(resultsMap); err != nil {
			log.Println("write:", err)
			break
		}

	}

}
