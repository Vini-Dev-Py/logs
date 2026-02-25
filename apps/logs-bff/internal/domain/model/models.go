package model

import "time"

type User struct {
	ID           string `json:"id"`
	CompanyID    string `json:"companyId"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	PasswordHash string `json:"-"`
}

type Annotation struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"nodeId"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}
