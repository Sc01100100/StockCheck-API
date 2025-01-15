package controllers

import (
	"log"

	"github.com/Sc01100100/SaveCash-API/module"
	"github.com/Sc01100100/SaveCash-API/models"
	"github.com/Sc01100100/SaveCash-API/config"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func GetItemsHandler(c *fiber.Ctx) error {
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

	rows, err := config.Database.Query(`
		SELECT id, user_id, name, description, stock, created_at 
		FROM items 
		WHERE user_id = $1
	`, intUserID)
	if err != nil {
		log.Printf("Error fetching items for user %d: %v\n", intUserID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch items",
		})
	}
	defer rows.Close()

	items := []models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Description, &item.Stock, &item.CreatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse items",
			})
		}
		items = append(items, item)
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"items":  items,
	})
}

func GetTransactionItemsHandler(c *fiber.Ctx) error {
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

	rows, err := config.Database.Query(`
		SELECT id, item_id, item_name, quantity, type, created_at, user_id 
		FROM stock_transactions 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, intUserID)
	if err != nil {
		log.Printf("Error fetching transactions for user %d: %v\n", intUserID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch transactions",
		})
	}
	defer rows.Close()

	transactions := []models.StockTransaction{}
	for rows.Next() {
		var transaction models.StockTransaction
		if err := rows.Scan(&transaction.ID, &transaction.ItemID, &transaction.ItemName, &transaction.Quantity, &transaction.Type, &transaction.CreatedAt, &transaction.UserID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse transactions",
			})
		}
		transactions = append(transactions, transaction)
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"transactions": transactions,
	})
}

func AddItemHandler(c *fiber.Ctx) error {
	item := new(models.Item)
	if err := c.BodyParser(item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	userID := c.Locals("user_id") 
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "UserID is missing in context",
		})
	}

	intUserID, ok := userID.(int)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid UserID format",
		})
	}

	_, err := config.Database.Exec(
		`INSERT INTO items (user_id, name, description, stock) VALUES ($1, $2, $3, $4)`,
		intUserID, item.Name, item.Description, item.Stock,
	)
	if err != nil {
		log.Printf("Error inserting item: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to add item",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "item added successfully",
	})
}

func RestockItemHandler(c *fiber.Ctx) error {
    userID := c.Locals("user_id")
    if userID == nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "UserID is missing in context",
        })
    }

    intUserID, ok := userID.(int)
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Invalid UserID format",
        })
    }

    itemID, err := strconv.Atoi(c.Params("id"))
    if err != nil || itemID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid item ID",
        })
    }

    var body struct {
        Quantity int `json:"quantity"`
    }
    if err := c.BodyParser(&body); err != nil || body.Quantity <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body or quantity must be greater than zero",
        })
    }

    if err := module.RestockItem(intUserID, itemID, body.Quantity); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Item restocked successfully",
    })
}

func SellItemHandler(c *fiber.Ctx) error {
    itemID, err := strconv.Atoi(c.Params("id"))
    if err != nil || itemID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid item ID",
        })
    }

    var body struct {
        Quantity int `json:"quantity"`
    }
    if err := c.BodyParser(&body); err != nil || body.Quantity <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body or quantity must be greater than zero",
        })
    }

    userID := c.Locals("user_id")
    if userID == nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "UserID is missing in context",
        })
    }

    intUserID, ok := userID.(int)
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Invalid UserID format",
        })
    }

    if err := module.SellItem(intUserID, itemID, body.Quantity); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Item sold successfully",
    })
}

func DeleteItemHandler(c *fiber.Ctx) error {
    userID := c.Locals("user_id")
    if userID == nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "UserID is missing in context",
        })
    }

    intUserID, ok := userID.(int)
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Invalid UserID format",
        })
    }

    itemID, err := strconv.Atoi(c.Params("id"))
    if err != nil || itemID <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid item ID",
        })
    }

    var ownerID int
    err = config.Database.QueryRow(`
        SELECT user_id FROM items WHERE id = $1
    `, itemID).Scan(&ownerID)
    if err != nil {
        if err.Error() == "sql: no rows in result set" {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Item not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch item details",
        })
    }

    if ownerID != intUserID {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "You are not authorized to delete this item",
        })
    }

    _, err = config.Database.Exec(`
        DELETE FROM items WHERE id = $1
    `, itemID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to delete item",
        })
    }

    return c.JSON(fiber.Map{
        "message": "Item deleted successfully",
    })
}