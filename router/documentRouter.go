package router

import (
	"github.com/gofiber/fiber/v2"
	controller "github.com/greyhands2/searchly/controllers"
)

var HandleDocumentRoutes = func(router fiber.Router) {
	router.Post("/", controller.InsertDocument)
	router.Get("/:query", controller.Search)
}
