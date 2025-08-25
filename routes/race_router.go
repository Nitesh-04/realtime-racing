package routes

import (
	"github.com/Nitesh-04/realtime-racing/controllers"
	"github.com/gofiber/fiber/v2"
)

func RaceRouter(api fiber.Router) {
	api.Post("/race/create", controllers.CreateRoom)
	api.Post("/race/join/:roomCode", controllers.JoinRoom)
	api.Post("/race/leave/:roomCode", controllers.LeaveRoom)
	api.Get("/race/:roomCode", controllers.GetRoomDetails)
	api.Post("/race/over/:roomCode", controllers.GameOver)
	api.Post("/race/updateResults/:roomCode", controllers.UpdateUserResult)
	api.Delete("/race/:roomCode", controllers.DeleteRoom)
}