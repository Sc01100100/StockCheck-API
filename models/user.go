package models

type User struct {
    ID       int     `json:"id"`
    Name     string  `json:"name"`
    Email    string  `json:"email"`
    Password string  `json:"password"`
    Role     string  `json:"role"`
    Balance  float64 `json:"balance"`
}