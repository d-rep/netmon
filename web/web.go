package web

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/drep/netmon/storage"
)

func Serve(port string, db *storage.Storage) error {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello world!")
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
