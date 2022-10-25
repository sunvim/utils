package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sunvim/utils/grace"
)

func main() {
	_, gsvc := grace.New(context.Background())

	svc := fiber.New()
	svc.Get("/:name", func(c *fiber.Ctx) error {
		log.Printf("hello %s \n", c.Params("name"))
		c.JSON(fmt.Sprintf("hello %s", c.Params("name")))
		return nil
	})

	gsvc.RegisterService("main service", func(ctx context.Context) error {
		log.Println("boot service")
		svc.Listen(":8080")
		return nil
	})

	gsvc.Wait()

}
