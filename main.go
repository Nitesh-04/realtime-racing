package main

import (
	"fmt"

	"github.com/Nitesh-04/realtime-racing/config"
	"github.com/Nitesh-04/realtime-racing/middleware"
	"github.com/Nitesh-04/realtime-racing/routes"
	"github.com/gofiber/fiber/v2"
	// "github.com/gofiber/fiber/v2/middleware/cors"
)

func init() {
	config.ConnectDB()
}

func main() {
	app := fiber.New()

	setupMiddlewares(app)
	setupRoutes(app)
	startServer(app)
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Use(middleware.CheckAuth())

	routes.AuthRouter(api)
}

func setupMiddlewares(app *fiber.App) {

	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins:     "*",
	// 	AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
	// 	AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Upgrade, Connection",
	// 	AllowCredentials: true,
	// 	}))
}

func startServer(app *fiber.App) {
	port := "8080"

	err := app.Listen(":" + port)

	fmt.Println("Server is running on port 8080")

	if err != nil {
		panic(err)
	}
}