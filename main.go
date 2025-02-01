package main

import (
	"log"
	"github.com/joho/godotenv"
	"github.com/gofiber/fiber/v2"
	"github.com/Sc01100100/SaveCash-API/config"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/Sc01100100/SaveCash-API/routes"
	_ "github.com/Sc01100100/SaveCash-API/docs"
)

// @title TEST SWAGGER SC
// @version 1.0
// @description This is a sample swagger for Fiber

// @contact.name API Support
// @contact.url github.com/Sc01100100/SaveCash-API
// @contact.email scsc01100100@gmail.com

// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	config.ConnectDB()
	defer config.Database.Close()

	log.Println("Database connection established.")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", 
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Content-Type, Authorization",
	}))

	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":8080"))
}