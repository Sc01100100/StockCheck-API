package test

import (
	"testing"

	"github.com/Sc01100100/SaveCash-API/module"
)

func TestInsertUser(t *testing.T) {
	name := "John Doe"
	email := "john.doe@example.com"
	password := "securepassword"
	role := "user" 

	insertedID, err := module.InsertUser(name, email, password, role)

	if err != nil {
		t.Errorf("Failed to insert user: %v", err)
		return
	}

	if insertedID <= 0 {
		t.Errorf("Failed to insert user: insertedID is invalid (%v)", insertedID)
		return
	}

	t.Logf("User inserted successfully with ID: %v", insertedID)
}
