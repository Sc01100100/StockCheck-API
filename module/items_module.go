package module

import (
	"fmt"
	"time"

	"github.com/Sc01100100/SaveCash-API/config"
	"github.com/Sc01100100/SaveCash-API/models"
)

func RestockItem(userID, itemID, quantity int) error {
    if quantity <= 0 {
        return fmt.Errorf("quantity must be greater than zero")
    }

    var item models.Item
    err := config.Database.QueryRow(`SELECT id, name, stock FROM items WHERE id = $1`, itemID).Scan(&item.ID, &item.Name, &item.Stock)
    if err != nil {
        return fmt.Errorf("failed to fetch item: %w", err)
    }

    item.Stock += quantity
    _, err = config.Database.Exec(`UPDATE items SET stock = $1 WHERE id = $2`, item.Stock, itemID)
    if err != nil {
        return fmt.Errorf("failed to update stock: %w", err)
    }

    stockTransaction := models.StockTransaction{
        ItemID:    itemID,
        ItemName:  item.Name, 
        Quantity:  quantity,
        Type:      "IN",
        CreatedAt: time.Now(),
        UserID:    userID,
    }

    _, err = config.Database.Exec(`
        INSERT INTO stock_transactions (item_id, item_name, quantity, type, created_at, user_id)
        VALUES ($1, $2, $3, $4, $5, $6)`,
        stockTransaction.ItemID, stockTransaction.ItemName, stockTransaction.Quantity,
        stockTransaction.Type, stockTransaction.CreatedAt, stockTransaction.UserID,
    )
    if err != nil {
        return fmt.Errorf("failed to record stock transaction: %w", err)
    }

    return nil
}

func SellItem(userID, itemID, quantity int) error {
    if quantity <= 0 {
        return fmt.Errorf("quantity must be greater than zero")
    }

    var item models.Item
    err := config.Database.QueryRow(`SELECT id, name, stock FROM items WHERE id = $1`, itemID).Scan(&item.ID, &item.Name, &item.Stock)
    if err != nil {
        return fmt.Errorf("failed to fetch item: %w", err)
    }

    if item.Stock < quantity {
        return fmt.Errorf("insufficient stock: available %d, required %d", item.Stock, quantity)
    }

    item.Stock -= quantity
    _, err = config.Database.Exec(`UPDATE items SET stock = $1 WHERE id = $2`, item.Stock, itemID)
    if err != nil {
        return fmt.Errorf("failed to update stock: %w", err)
    }

    stockTransaction := models.StockTransaction{
        ItemID:    itemID,
        ItemName:  item.Name, 
        Quantity:  -quantity,
        Type:      "OUT",
        CreatedAt: time.Now(),
        UserID:    userID,
    }

    _, err = config.Database.Exec(`
        INSERT INTO stock_transactions (item_id, item_name, quantity, type, created_at, user_id)
        VALUES ($1, $2, $3, $4, $5, $6)`,
        stockTransaction.ItemID, stockTransaction.ItemName, stockTransaction.Quantity,
        stockTransaction.Type, stockTransaction.CreatedAt, stockTransaction.UserID,
    )
    if err != nil {
        return fmt.Errorf("failed to record stock transaction: %w", err)
    }

    return nil
}