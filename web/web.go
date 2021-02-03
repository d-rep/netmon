package web

import "github.com/gofiber/fiber/v2"

func Serve(port string) error {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello world!")
	})

	return app.Listen(":" + port)
}
