package controllers

import (
	"strconv"
	"strings"
	"log"
	"regexp"

	"github.com/Sc01100100/SaveCash-API/module"
	"github.com/Sc01100100/SaveCash-API/models"
	"github.com/Sc01100100/SaveCash-API/config"
	"github.com/Sc01100100/SaveCash-API/utils"
	"github.com/gofiber/fiber/v2"
)

func GetAllUser(c *fiber.Ctx) error {
	users := module.GetAllUsers()

	if len(users) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "No users found",
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Users retrieved successfully",
		"data":    users,
	})
}

func isValidEmail(email string) bool {
	var emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || char == '^' || char == '&' || char == '*' || char == '(' || char == ')' || char == '_' || char == '+' || char == '-' || char == '=' || char == '{' || char == '}' || char == '[' || char == ']' || char == ':' || char == ';' || char == '"' || char == '\'' || char == '<' || char == '>' || char == ',' || char == '.' || char == '?' || char == '/' || char == '~':
			hasSpecial = true
		}
	}

	return hasUpper && hasNumber && hasSpecial
}

func InsertUser(c *fiber.Ctx) error {
	type RequestBody struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	var body RequestBody

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"data":    nil,
		})
	}

	if body.Name == "" || body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "All fields (name, email, password) are required",
			"data":    nil,
		})
	}

	if !isValidEmail(body.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid email format.",
			"data":    nil,
		})
	}

	if !isValidPassword(body.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Password must be at least 8 characters, with an uppercase letter, number, and special character.",
			"data":    nil,
		})
	}

	if body.Role == "" {
		body.Role = "user"
	}

	insertedID, err := module.InsertUser(body.Name, body.Email, body.Password, body.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "User created successfully",
		"data":    map[string]interface{}{"id": insertedID},
	})
}

func LoginUser(c *fiber.Ctx) error {
	type RequestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body RequestBody

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"data":    nil,
		})
	}

	userID, role, err := module.LoginUser(body.Email, body.Password)
	if err != nil {
		status := fiber.StatusUnauthorized
		if err.Error() == "user not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
	}

	token, err := utils.GenerateJWT(strconv.Itoa(userID), role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate token",
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Login successful",
		"data": map[string]interface{}{
			"user_id": userID,
			"role":    role,
			"token":   token,
		},
	})
}

func LogoutUser(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Authorization token is missing",
		})
	}

	token = strings.Replace(token, "Bearer ", "", 1)

	query := `INSERT INTO token_blacklist (token) VALUES ($1)`
	_, err := config.Database.Exec(query, token)
	if err != nil {
		log.Printf("Error blacklisting token: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to logout",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Logout successful",
	})
}

func GetUserInfo(c *fiber.Ctx) error {
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

	var user models.User
	query := `SELECT name, balance FROM users WHERE id = $1`
	err := config.Database.QueryRow(query, intUserID).Scan(&user.Name, &user.Balance)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve user information",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"name":    user.Name,
			"balance": user.Balance,
		},
	})
}