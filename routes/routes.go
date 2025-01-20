package routes

import (
	"github.com/Sc01100100/SaveCash-API/controllers"
	"github.com/Sc01100100/SaveCash-API/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/savecash")

	api.Post("/register", controllers.InsertUser)
	api.Post("/login", controllers.LoginUser)
	api.Post("/logout", controllers.LogoutUser)

	protected := api.Group("/", middlewares.AuthMiddleware())

	protected.Post("/transactions", controllers.CreateTransactionHandler)
	protected.Get("/transactions", controllers.GetTransactionsHandler)
	protected.Get("/transactions/:id", controllers.GetTransactionByIDHandler)
	protected.Put("/transactions/:id", controllers.UpdateTransactionHandler) 
	protected.Delete("/transactions/:id", controllers.DeleteTransactionHandler) 

	protected.Post("/incomes", controllers.CreateIncomeHandler)
	protected.Get("/incomes", controllers.GetIncomesHandler)
	protected.Get("/incomes/:id", controllers.GetIncomeByIDHandler)
	protected.Put("/incomes/:id", controllers.UpdateIncomeHandler)
	protected.Delete("/incomes/:id", controllers.DeleteIncomeHandler)

	protected.Post("/items", controllers.AddItemHandler)            
	protected.Get("/items", controllers.GetItemsHandler)   
	protected.Get("/txitems", controllers.GetTransactionItemsHandler)           
	protected.Put("/items/restock/:id", controllers.RestockItemHandler)
	protected.Put("/items/sell/:id", controllers.SellItemHandler)
	protected.Delete("/items/:id", controllers.DeleteItemHandler)  

	protected.Get("/user/info", controllers.GetUserInfo)

	admin := protected.Group("/admin") 
	admin.Use(middlewares.AdminMiddleware()) 
	admin.Get("/users", controllers.GetAllUser) 
}