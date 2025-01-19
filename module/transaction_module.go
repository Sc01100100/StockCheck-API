package module

import (
	"fmt"
	"log"
	"time"

	"github.com/Sc01100100/SaveCash-API/config"
	"github.com/Sc01100100/SaveCash-API/models"
)

func CreateTransaction(userID int, amount float64, category, description string) (models.Transaction, error) {
	queryIncome := `SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE user_id = $1`
	var totalIncome float64
	err := config.Database.QueryRow(queryIncome, userID).Scan(&totalIncome)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to fetch total income: %w", err)
	}
	log.Printf("Total Income for UserID %d: %.2f\n", userID, totalIncome)

	queryExpense := `SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1`
	var totalExpense float64
	err = config.Database.QueryRow(queryExpense, userID).Scan(&totalExpense)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to fetch total expenses: %w", err)
	}
	log.Printf("Total Expenses for UserID %d: %.2f\n", userID, totalExpense)

	availableBalance := totalIncome - totalExpense
	log.Printf("Available Balance for UserID %d: %.2f\n", userID, availableBalance)

	if amount > availableBalance {
		return models.Transaction{}, fmt.Errorf("insufficient funds: available %.2f, required %.2f", availableBalance, amount)
	}

	query := `
		INSERT INTO transactions (user_id, amount, category, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, amount, category, description, created_at
	`
	var transaction models.Transaction
	err = config.Database.QueryRow(query, userID, amount, category, description, time.Now()).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Category, &transaction.Description, &transaction.CreatedAt,
	)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to create transaction: %w", err)
	}

	_, err = config.Database.Exec(`UPDATE users SET balance = balance - $1 WHERE id = $2`, amount, userID)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to update user balance after transaction: %w", err)
	}

	return transaction, nil
}

func CreateIncome(userID int, amount float64, source string) (models.Income, error) {
	var exists bool
	err := config.Database.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking user existence: %v\n", err)
		return models.Income{}, fmt.Errorf("failed to check user existence")
	}
	if !exists {
		log.Printf("User with ID %d does not exist\n", userID)
		return models.Income{}, fmt.Errorf("user with ID %d does not exist", userID)
	}

	log.Printf("Inserting income: UserID: %d, Amount: %.2f, Source: %s\n", userID, amount, source)

	query := `
		INSERT INTO incomes (user_id, amount, source, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, amount, source, created_at
	`
	var income models.Income
	err = config.Database.QueryRow(query, userID, amount, source, time.Now()).Scan(&income.ID, &income.UserID, &income.Amount, &income.Source, &income.CreatedAt)
	if err != nil {
		log.Printf("Error creating income: %v\n", err)
		return income, err
	}

	log.Printf("Income created successfully: ID: %d, UserID: %d, Amount: %.2f, Source: %s, CreatedAt: %s\n",
		income.ID, income.UserID, income.Amount, income.Source, income.CreatedAt)

	_, err = config.Database.Exec(`UPDATE users SET balance = balance + $1 WHERE id = $2`, amount, userID)
	if err != nil {
		log.Printf("Error updating user balance: %v\n", err)
		return models.Income{}, fmt.Errorf("failed to update user balance")
	}

	return income, nil
}

func GetTransactions(userID int) ([]models.Transaction, error) {
	query := `SELECT id, user_id, amount, category, description, created_at FROM transactions WHERE user_id = $1`
	rows, err := config.Database.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Category, &transaction.Description, &transaction.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func GetIncomes(userID int) ([]models.Income, error) {
	query := `SELECT id, user_id, amount, source, created_at FROM incomes WHERE user_id = $1`
	rows, err := config.Database.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch incomes: %w", err)
	}
	defer rows.Close()

	var incomes []models.Income
	for rows.Next() {
		var income models.Income
		if err := rows.Scan(&income.ID, &income.UserID, &income.Amount, &income.Source, &income.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan income: %w", err)
		}
		incomes = append(incomes, income)
	}

	return incomes, nil
}

func DeleteTransaction(id int) error {
    _, err := config.Database.Exec(`DELETE FROM transactions WHERE id = $1`, id)
    if err != nil {
        return fmt.Errorf("failed to delete transaction: %w", err)
    }
    return nil
}

func UpdateTransaction(transactionID int, userID int, amount float64, category string, description string) (*models.Transaction, error) {
    var existingTransaction models.Transaction
    err := config.Database.QueryRow(`SELECT id, user_id, amount, category, description FROM transactions WHERE id = $1`, transactionID).Scan(&existingTransaction.ID, &existingTransaction.UserID, &existingTransaction.Amount, &existingTransaction.Category, &existingTransaction.Description)
    if err != nil {
        return nil, fmt.Errorf("transaction not found: %w", err)
    }

    if existingTransaction.UserID != userID {
        return nil, fmt.Errorf("you are not authorized to update this transaction")
    }

    _, err = config.Database.Exec(`UPDATE transactions SET amount = $1, category = $2, description = $3 WHERE id = $4`, amount, category, description, transactionID)
    if err != nil {
        return nil, fmt.Errorf("failed to update transaction: %w", err)
    }

    existingTransaction.Amount = amount
    existingTransaction.Category = category
    existingTransaction.Description = description
    return &existingTransaction, nil
}