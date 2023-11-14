package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/greyhands2/searchly/config"
	"github.com/greyhands2/searchly/controllers"
	router "github.com/greyhands2/searchly/router"
)

func main() {
	es, err := config.ElasticSearchConnect()
	if err != nil {
		log.Fatalf("Error creating Elasticsearch Client %s", err)

	}

	app := fiber.New()

	//fiber middleware to set the es connection to context that is compatible with fiber
	app.Use(func(c *fiber.Ctx) error {
		//add the es connection to app context
		ctx := context.WithValue(c.Context(), "esConnection", es)
		c.Locals("ctx", ctx)
		return c.Next()
	})

	app.Use(logger.New())
	//handle panics if any
	app.Use(recover.New())
	app.Use("/socket", func(c *fiber.Ctx) error {

		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		var upgrade bool = websocket.IsWebSocketUpgrade(c)
		fmt.Println(upgrade)
		if upgrade {
			fmt.Println("upgraded")

			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	//
	app.Get("/socket", websocket.New(controllers.SearchSocketServer))

	//let us create an app group
	api := app.Group("/api")

	router.HandleDocumentRoutes(api.Group("/document")) //used an alias named router on the import

	//listen to server on port 3000
	log.Fatal(app.Listen(":3000"))

}
