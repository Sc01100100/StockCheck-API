package controllers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Sc01100100/SaveCash-API/models"
	"github.com/Sc01100100/SaveCash-API/module"
	"github.com/Sc01100100/SaveCash-API/config"
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

    if body.Amount <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Amount must be greater than zero",
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

func DeleteTransactionHandler(c *fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil || id <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid transaction ID",
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
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid UserID format",
        })
    }

    var ownerID int
    err = config.Database.QueryRow(`SELECT user_id FROM transactions WHERE id = $1`, id).Scan(&ownerID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "status":  "error",
            "message": "Transaction not found",
        })
    }

    if ownerID != intUserID {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "status":  "error",
            "message": "You are not authorized to delete this transaction",
        })
    }

    err = module.DeleteTransaction(id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "status":  "success",
        "message": "Transaction deleted successfully",
    })
}

func UpdateTransactionHandler(c *fiber.Ctx) error {
    transactionID, err := strconv.Atoi(c.Params("id"))
    if err != nil || transactionID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid transaction ID",
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
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid UserID format",
        })
    }

    var body struct {
        Amount      float64 `json:"amount"`
        Category    string  `json:"category"`
        Description string  `json:"description"`
    }
    if err := c.BodyParser(&body); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid request body",
        })
    }

    updatedTransaction, err := module.UpdateTransaction(transactionID, intUserID, body.Amount, body.Category, body.Description)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "status":      "success",
        "transaction": updatedTransaction,
    })
}

func DeleteIncomeHandler(c *fiber.Ctx) error {
    incomeID, err := strconv.Atoi(c.Params("id"))
    if err != nil || incomeID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid income ID",
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
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid UserID format",
        })
    }

    err = module.DeleteIncome(incomeID, intUserID) 
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "status":  "success",
        "message": "Income deleted successfully",
    })
}

func UpdateIncomeHandler(c *fiber.Ctx) error {
    incomeID, err := strconv.Atoi(c.Params("id"))
    if err != nil || incomeID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid income ID",
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
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid UserID format",
        })
    }

    var body struct {
        Amount float64 `json:"amount"`
        Source string  `json:"source"`
    }
    if err := c.BodyParser(&body); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid request body",
        })
    }

    updatedIncome, err := module.UpdateIncome(incomeID, intUserID, body.Amount, body.Source)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "status":  "success",
        "income":  updatedIncome,
    })
}

func GetIncomeByIDHandler(c *fiber.Ctx) error {
    incomeID, err := strconv.Atoi(c.Params("id"))
    if err != nil || incomeID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid income ID",
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
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid UserID format",
        })
    }

    income, err := module.GetIncomeByID(incomeID, intUserID) 
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "income": income,
    })
}

func GetTransactionByIDHandler(c *fiber.Ctx) error {
    transactionID, err := strconv.Atoi(c.Params("id"))
    if err != nil || transactionID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid transaction ID",
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
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid UserID format",
        })
    }

    transaction, err := module.GetTransactionByID(transactionID, intUserID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "transaction": transaction,
    })
}