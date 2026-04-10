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
	err := r.DB.QueryRow(ctx, `
		SELECT u.id::text, u.company_id::text, u.name, u.email, r.name, u.password_hash 
		FROM users u 
		LEFT JOIN roles r ON u.role_id = r.id 
		WHERE u.email=$1`, email).
		Scan(&u.ID, &u.CompanyID, &u.Name, &u.Email, &u.Role, &u.PasswordHash)
		
	if err == nil {
		rows, _ := r.DB.Query(ctx, "SELECT permission_name FROM role_permissions rp JOIN users u ON u.role_id = rp.role_id WHERE u.id = $1", u.ID)
		defer rows.Close()
		for rows.Next() {
			var p string
			_ = rows.Scan(&p)
			u.Permissions = append(u.Permissions, p)
		}
	}
	return u, err
}

func (r Repositories) FindByID(ctx context.Context, id string) (model.User, error) {
	var u model.User
	err := r.DB.QueryRow(ctx, `
		SELECT u.id::text, u.company_id::text, u.name, u.email, r.name, u.password_hash 
		FROM users u 
		LEFT JOIN roles r ON u.role_id = r.id 
		WHERE u.id=$1`, id).
		Scan(&u.ID, &u.CompanyID, &u.Name, &u.Email, &u.Role, &u.PasswordHash)
		
	if err == nil {
		rows, _ := r.DB.Query(ctx, "SELECT permission_name FROM role_permissions rp JOIN users u ON u.role_id = rp.role_id WHERE u.id = $1", id)
		defer rows.Close()
		for rows.Next() {
			var p string
			_ = rows.Scan(&p)
			u.Permissions = append(u.Permissions, p)
		}
	}
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

type ListedUser struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Role  *string `json:"role"`
}

func (r Repositories) ListUsers(ctx context.Context, companyID string) ([]ListedUser, error) {
	rows, err := r.DB.Query(ctx, "SELECT u.id::text, u.name, u.email, r.name FROM users u LEFT JOIN roles r ON u.role_id = r.id WHERE u.company_id=$1", companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ListedUser
	for rows.Next() {
		var u ListedUser
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err == nil {
			out = append(out, u)
		}
	}
	return out, nil
}

type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r Repositories) ListRoles(ctx context.Context, companyID string) ([]Role, error) {
	rows, err := r.DB.Query(ctx, "SELECT id::text, name, description FROM roles WHERE company_id IS NULL OR company_id=$1", companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err == nil {
			out = append(out, role)
		}
	}
	return out, nil
}

func (r Repositories) UpdateUserRole(ctx context.Context, companyID, userID, roleID string) error {
	_, err := r.DB.Exec(ctx, "UPDATE users SET role_id=$1 WHERE id=$2 AND company_id=$3", roleID, userID, companyID)
	return err
}
