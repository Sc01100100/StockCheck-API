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
	rows, err := config.Database.Query(`SELECT id, user_id, name, description, stock, created_at FROM items`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch items",
		})
	}
	defer rows.Close()

	items := []models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Description, &item.Stock, &item.CreatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to parse items",
			})
		}
		items = append(items, item)
	}

	return c.JSON(items)
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
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid item ID",
        })
    }

    body := struct {
        Quantity int `json:"quantity"`
    }{}

    if err := c.BodyParser(&body); err != nil || body.Quantity <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body or quantity must be greater than zero",
        })
    }

    err = module.RestockItem(intUserID, itemID, body.Quantity)
    if err != nil {
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
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid item ID",
        })
    }

    type Request struct {
        Quantity int `json:"quantity"`
    }
    req := new(Request)

    if err := c.BodyParser(req); err != nil || req.Quantity <= 0 {
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

    err = module.SellItem(intUserID, itemID, req.Quantity)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Item sold successfully",
    })
}