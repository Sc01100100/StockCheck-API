package module

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"errors"

	"github.com/Sc01100100/SaveCash-API/config"
	"github.com/Sc01100100/SaveCash-API/models"
	"golang.org/x/crypto/bcrypt"
)

func InsertUser(name, email, password, role string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error encrypting password: %v\n", err)
		return 0, fmt.Errorf("failed to encrypt password")
	}

	newUser := models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	query := `
		INSERT INTO users (name, email, password, role)
		VALUES ($1, $2, $3, $4) RETURNING id
	`

	var insertedID int

	err = config.Database.QueryRow(query, newUser.Name, newUser.Email, newUser.Password, newUser.Role).Scan(&insertedID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			log.Printf("InsertUser error: email %s already exists\n", newUser.Email)
			return 0, fmt.Errorf("email already exists")
		}

		log.Printf("InsertUser error: %v\n", err)
		return 0, fmt.Errorf("failed to insert user")
	}

	log.Printf("Inserted new user with ID: %d\n", insertedID)
	return insertedID, nil
}


func GetAllUsers() []models.User {
	query := `SELECT id, name, email, role FROM users`

	rows, err := config.Database.Query(query)
	if err != nil {
		log.Printf("GetAllUsers error: %v\n", err)
		return nil
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role)
		if err != nil {
			log.Printf("Row scan error: %v\n", err)
			continue
		}
		users = append(users, user)
	}

	return users
}

func LoginUser(email, password string) (int, string, error) {
	var user models.User

	query := `SELECT id, password, role FROM users WHERE email = $1`
	err := config.Database.QueryRow(query, email).Scan(&user.ID, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Login failed: user with email %s not found\n", email)
			return 0, "", fmt.Errorf("user not found")
		}
		log.Printf("Database error during login for email %s: %v\n", email, err)
		return 0, "", fmt.Errorf("database error")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Printf("Login failed: invalid password for email %s\n", email)
		return 0, "", fmt.Errorf("invalid password")
	}


	log.Printf("User %d logged in successfully with role: %s\n", user.ID, user.Role)
	return user.ID, user.Role, nil
}