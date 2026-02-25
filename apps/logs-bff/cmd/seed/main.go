package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := pgxpool.New(context.Background(), env("DATABASE_URL", "postgres://logs:logs@localhost:5432/logs?sslmode=disable"))
	if err != nil { panic(err) }
	defer db.Close()

	companyID := uuid.NewString()
	userID := uuid.NewString()
	apiKey := "logs_dev_api_key"
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	_, _ = db.Exec(context.Background(), "INSERT INTO companies(id,name,api_key) VALUES($1,$2,$3) ON CONFLICT (api_key) DO NOTHING", companyID, "Logs Demo", apiKey)
	_, _ = db.Exec(context.Background(), "INSERT INTO users(id,company_id,name,email,password_hash,role) VALUES($1,(SELECT id FROM companies WHERE api_key=$2),$3,$4,$5,$6) ON CONFLICT (email) DO NOTHING", userID, apiKey, "Admin", "admin@logs.local", string(hash), "ADMIN")
	fmt.Println("seed complete")
	fmt.Println("email=admin@logs.local password=admin123 apiKey=logs_dev_api_key")
}

func env(k, d string) string { if v:=os.Getenv(k); v!="" { return v }; return d }
