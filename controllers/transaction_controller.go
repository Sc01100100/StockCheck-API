package controllers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Sc01100100/SaveCash-API/models"
	"github.com/Sc01100100/SaveCash-API/module"
	"github.com/gofiber/fiber/v2"
)

func CreateTransactionHandler(c *fiber.Ctx) error {
	type RequestBody struct {
		Amount      float64 `json:"amount"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
	}

	var body RequestBody

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "UserID is missing in context",
		})
	}

	intUserID, ok := userID.(int)
	if !ok || intUserID == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid UserID format",
		})
	}

	transaction, err := module.CreateTransaction(intUserID, body.Amount, body.Category, body.Description)
	if err != nil {
		if err.Error() == fmt.Sprintf("insufficient funds: available %.2f, required %.2f", 0.0, body.Amount) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":      "success",
		"transaction": transaction,
	})
}

func DeleteTransactionHandler(c *fiber.Ctx) error {
	transactionID, err := strconv.Atoi(c.Params("transactionID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transaction ID",
		})
	}

	if err := module.DeleteTransaction(transactionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete transaction",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Transaction deleted successfully",
	})
}

func CreateIncomeHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		log.Println("UserID is missing in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "UserID is missing in context",
		})
	}

	intUserID, ok := userID.(int)
	if !ok || intUserID == 0 {
		log.Printf("Invalid UserID from context: %v\n", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid UserID format",
		})
	}

	log.Printf("Creating income for UserID: %d\n", intUserID)

	var income models.Income
	if err := c.BodyParser(&income); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid income data",
		})
	}

	if income.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Amount must be greater than zero",
		})
	}
	if income.Source == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Source cannot be empty",
		})
	}

	newIncome, err := module.CreateIncome(intUserID, income.Amount, income.Source)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create income",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"income": newIncome,
	})
}

func DeleteIncomeHandler(c *fiber.Ctx) error {
	incomeID, err := strconv.Atoi(c.Params("incomeID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid income ID",
		})
	}

	if err := module.DeleteIncome(incomeID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete income",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Income deleted successfully",
	})
}

func GetTransactionsHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		log.Println("UserID is missing in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "UserID is missing in context",
		})
	}

	intUserID, ok := userID.(int)
	if !ok || intUserID == 0 {
		log.Printf("Invalid UserID from context: %v\n", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid UserID format",
		})
	}

	transactions, err := module.GetTransactions(intUserID)
	if err != nil {
		log.Printf("Error fetching transactions: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch transactions",
		})
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"transactions": transactions,
	})
}

func GetIncomesHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		log.Println("UserID is missing in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "UserID is missing in context",
		})
	}

	intUserID, ok := userID.(int)
	if !ok || intUserID == 0 {
		log.Printf("Invalid UserID from context: %v\n", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid UserID format",
		})
	}

	incomes, err := module.GetIncomes(intUserID)
	if err != nil {
		log.Printf("Error fetching incomes: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch incomes",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"incomes": incomes,
	})
}