package test

import (
	"testing"
	"os"

	"github.com/Sc01100100/SaveCash-API/module"
	"github.com/Sc01100100/SaveCash-API/config"
)

func TestMain(m *testing.M) {
	config.ConnectDB()
	code := m.Run()
	defer config.Database.Close()
	os.Exit(code)
}
func TestCreateIncome(t *testing.T) {
	userID := 1
	amount := 1000.0
	source := "Salary"

	createdIncome, err := module.CreateIncome(userID, amount, source)
	if err != nil {
		t.Errorf("Failed to create income: %v", err)
	} else {
		if createdIncome.ID == 0 {
			t.Errorf("Failed to create income: ID is 0")
		} else {
			t.Logf("Income created successfully with ID: %v", createdIncome.ID)
		}

		if createdIncome.UserID != userID {
			t.Errorf("Expected UserID %d, but got %d", userID, createdIncome.UserID)
		}
		if createdIncome.Amount != amount {
			t.Errorf("Expected Amount %.2f, but got %.2f", amount, createdIncome.Amount)
		}
		if createdIncome.Source != source {
			t.Errorf("Expected Source %s, but got %s", source, createdIncome.Source)
		}
	}
}
func TestCreateTransaction(t *testing.T) {
	userID := 1
	amount := 500.0
	category := "buy car"
	description := "buy car for "

	transaction, err := module.CreateTransaction(userID, amount, category, description)
	if err != nil {
		t.Errorf("Failed to create transaction: %v", err)
	} else {
		if transaction.ID == 0 {
			t.Errorf("Transaction ID is 0")
		}
		if transaction.Category != category {
			t.Errorf("Expected category '%s', got '%s'", category, transaction.Category)
		}
		t.Logf("Transaction created with ID: %d, Category: %s", transaction.ID, transaction.Category)
	}

	amount = -100.0
	_, err = module.CreateTransaction(userID, amount, category, description)
	if err == nil {
		t.Errorf("Expected error for negative amount, but got none")
	} else {
		t.Logf("Error for negative amount as expected: %v", err)
	}

	category = ""
	_, err = module.CreateTransaction(userID, amount, category, description)
	if err == nil {
		t.Errorf("Expected error for empty category, but got none")
	} else {
		t.Logf("Error for empty category as expected: %v", err)
	}
}
