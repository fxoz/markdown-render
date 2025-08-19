package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	pages := SetupFiles()

	for path := range pages {
		fmt.Println("Registered Path:", path)
	}

	app.Static("/_static", "./static")
	app.Get("*", func(c *fiber.Ctx) error {
		fmt.Println("Requested Path:", c.Path())

		if _, exists := pages[c.Path()]; !exists {
			return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
		}

		html := pages[c.Path()]

		return c.Type("html").SendString(html)
	})

	log.Fatal(app.Listen("localhost:3002"))
}
