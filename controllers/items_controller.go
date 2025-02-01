package controllers

import (
	"log"

	"github.com/Sc01100100/SaveCash-API/module"
	"github.com/Sc01100100/SaveCash-API/models"
	"github.com/Sc01100100/SaveCash-API/config"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

// @Summary Get all items for the authenticated user
// @Description This endpoint fetches all items associated with the currently authenticated user, including item details like name, description, stock, and created date.
// @Tags Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /savecash/items [get]
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

// @Summary Get all transactions for items of the authenticated user
// @Description This endpoint fetches all stock transactions related to items for the currently authenticated user, including details like item name, quantity, type, and created date.
// @Tags Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /savecash/txitems [get]
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

// @Summary Add a new item
// @Description This endpoint allows a user to add a new item to the inventory. The item requires a name, description, and stock count. The stock must be greater than zero.
// @Tags Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param item body models.Item true "Item data"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /savecash/items [post]
func AddItemHandler(c *fiber.Ctx) error {
	item := new(models.Item)
	if err := c.BodyParser(item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if item.Stock <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Stock must be greater than zero",
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

	item.UserID = intUserID

	_, err := config.Database.Exec(
		`INSERT INTO items (user_id, name, description, stock) VALUES ($1, $2, $3, $4)`,
		item.UserID, item.Name, item.Description, item.Stock,
	)
	if err != nil {
		log.Printf("Error inserting item: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add item",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Item added successfully",
	})
}

// @Summary Restock an item for the authenticated user
// @Description This endpoint allows the authenticated user to restock an item they own. The user must provide the item ID and the quantity to restock. The item must belong to the user making the request.
// @Tags Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Item ID"
// @Param stock body models.StockTransaction true "Stock"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /savecash/items/restock/{id} [put]
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

	var ownerID int
	err = config.Database.QueryRow(`SELECT user_id FROM items WHERE id = $1`, itemID).Scan(&ownerID)
	if err != nil || ownerID != intUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You do not have permission to modify this item",
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

// @Summary Sell an item for the authenticated user
// @Description This endpoint allows the authenticated user to sell an item they own. The user must provide the item ID and the quantity to sell. The item must belong to the user making the request.
// @Tags Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Item ID"
// @Param stock body models.StockTransaction true "Stock"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /savecash/items/sell/{id} [put]
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

	var ownerID int
	err = config.Database.QueryRow(`SELECT user_id FROM items WHERE id = $1`, itemID).Scan(&ownerID)
	if err != nil || ownerID != intUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You do not have permission to modify this item",
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

// @Summary Delete an item for the authenticated user
// @Description This endpoint allows the authenticated user to delete an item they own. The user must provide the item ID to be deleted. The item must belong to the user making the request.
// @Tags Items
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Item ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /savecash/items/{id} [delete]
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