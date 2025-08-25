package main

import (
	"fmt"

	"github.com/Nitesh-04/realtime-racing/config"
	"github.com/Nitesh-04/realtime-racing/middleware"
	"github.com/Nitesh-04/realtime-racing/routes"
	"github.com/Nitesh-04/realtime-racing/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	// "github.com/gofiber/fiber/v2/middleware/cors"
)

func init() {
	config.ConnectDB()
}

func main() {
	app := fiber.New()

	setupMiddlewares(app)
	setupRoutes(app)
	setupWebSocketRoutes(app)
	startServer(app)
}

func setupMiddlewares(app *fiber.App) {

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "*",
		}))
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Use(middleware.CheckAuth())

	routes.MainRouter(api)
}

func setupWebSocketRoutes(app *fiber.App) {
	app.Use(cors.New())
    app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:room_code", websocket.New(func(c *websocket.Conn) {
		websockets.Hub.HandleConnection(c)
	}))
}


func startServer(app *fiber.App) {
	port := "8080"

	err := app.Listen(":" + port)

	fmt.Println("Server is running on port 8080")

	if err != nil {
		panic(err)
	}
}