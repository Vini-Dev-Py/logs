package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyRepo struct{ DB *pgxpool.Pool }

func (r CompanyRepo) CompanyIDByAPIKey(ctx context.Context, apiKey string) (string, error) {
	var companyID string
	err := r.DB.QueryRow(ctx, "SELECT id::text FROM companies WHERE api_key=$1", apiKey).Scan(&companyID)
	return companyID, err
}
