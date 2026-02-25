package postgres

import (
	"context"
	"logs-bff/internal/domain/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct{ DB *pgxpool.Pool }

func (r Repositories) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.DB.QueryRow(ctx, "SELECT id::text, company_id::text, name, email, role, password_hash FROM users WHERE email=$1", email).
		Scan(&u.ID, &u.CompanyID, &u.Name, &u.Email, &u.Role, &u.PasswordHash)
	return u, err
}

func (r Repositories) FindByID(ctx context.Context, id string) (model.User, error) {
	var u model.User
	err := r.DB.QueryRow(ctx, "SELECT id::text, company_id::text, name, email, role, password_hash FROM users WHERE id=$1", id).
		Scan(&u.ID, &u.CompanyID, &u.Name, &u.Email, &u.Role, &u.PasswordHash)
	return u, err
}

func (r Repositories) ListByTrace(ctx context.Context, companyID, traceID string) ([]model.Annotation, error) {
	rows, err := r.DB.Query(ctx, "SELECT id::text,node_id,x,y,text,created_at FROM annotations WHERE company_id=$1 AND trace_id=$2", companyID, traceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Annotation{}
	for rows.Next() {
		var a model.Annotation
		if err := rows.Scan(&a.ID, &a.NodeID, &a.X, &a.Y, &a.Text, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func (r Repositories) Create(ctx context.Context, companyID, userID, traceID, nodeID string, x, y float64, text string) (string, error) {
	id := uuid.NewString()
	_, err := r.DB.Exec(ctx, "INSERT INTO annotations(id,company_id,trace_id,node_id,x,y,text,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", id, companyID, traceID, nodeID, x, y, text, userID)
	return id, err
}

func (r Repositories) Update(ctx context.Context, companyID, id, text string, x, y float64) error {
	_, err := r.DB.Exec(ctx, "UPDATE annotations SET text=$1,x=$2,y=$3 WHERE id=$4 AND company_id=$5", text, x, y, id, companyID)
	return err
}

func (r Repositories) Delete(ctx context.Context, companyID, id string) error {
	_, err := r.DB.Exec(ctx, "DELETE FROM annotations WHERE id=$1 AND company_id=$2", id, companyID)
	return err
}
