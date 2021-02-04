package web

import (
	"log"

	"github.com/gofiber/template/html"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/drep/netmon/storage"
)

func Serve(port string, db *storage.Storage) error {
	engine := html.New("./views", ".html")
	engine.Reload(true)
	engine.Debug(true)
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		calls, err := db.GetRecentCalls(10)
		if err != nil {
			log.Println(err)
		}
		return c.Render("index", fiber.Map{
			"Title": "Network Monitor",
			"calls": calls,
		})
	})

	app.Get("/status", func(c *fiber.Ctx) error {
		calls, err := db.GetRecentCalls(10)
		if err != nil {
			log.Println(err)
		}
		return c.JSON(calls)
	})

	return app.Listen("localhost:" + port)
}
