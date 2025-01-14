package main

import (
	"log"
	"github.com/joho/godotenv"
	"github.com/gofiber/fiber/v2"
	"github.com/Sc01100100/SaveCash-API/config"
	"github.com/Sc01100100/SaveCash-API/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	config.ConnectDB()
	defer config.Database.Close()

	log.Println("Database connection established.")

	app := fiber.New()

	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":8080"))
}