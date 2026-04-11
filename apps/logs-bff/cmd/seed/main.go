package main

import (
	"context"
	"fmt"
	"os"
	"time"

	search "shared-search"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := pgxpool.New(context.Background(), env("DATABASE_URL", "postgres://logs:logs@localhost:5432/logs?sslmode=disable"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	companyID := uuid.NewString()
	userID := uuid.NewString()
	apiKey := "logs_dev_api_key"
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	_, _ = db.Exec(context.Background(), "INSERT INTO companies(id,name,api_key) VALUES($1,$2,$3) ON CONFLICT (api_key) DO NOTHING", companyID, "Logs Demo", apiKey)
	
	// Pre-create Default Role
	roleID := uuid.NewString()
	_, _ = db.Exec(context.Background(), "INSERT INTO roles(id,company_id,name,description,is_system_default) VALUES($1,(SELECT id FROM companies WHERE api_key=$2),$3,$4,$5) ON CONFLICT DO NOTHING", roleID, apiKey, "Administrador", "Acesso total a todas as funções", true)
	_, _ = db.Exec(context.Background(), "INSERT INTO role_permissions(role_id,permission_name) VALUES($1,$2) ON CONFLICT DO NOTHING", roleID, "traces:read")
	_, _ = db.Exec(context.Background(), "INSERT INTO role_permissions(role_id,permission_name) VALUES($1,$2) ON CONFLICT DO NOTHING", roleID, "annotations:write")
	_, _ = db.Exec(context.Background(), "INSERT INTO role_permissions(role_id,permission_name) VALUES($1,$2) ON CONFLICT DO NOTHING", roleID, "users:manage")
	
	// Create user attached to new role (ignoring old text role)
	_, _ = db.Exec(context.Background(), "INSERT INTO users(id,company_id,name,email,password_hash,role,role_id) VALUES($1,(SELECT id FROM companies WHERE api_key=$2),$3,$4,$5,$6,$7) ON CONFLICT (email) DO NOTHING", userID, apiKey, "Admin", "admin@logs.local", string(hash), "ADMIN", roleID)

	// Create cassandra connection
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "logs"
	session, err := cluster.CreateSession()
	if err == nil {
		defer session.Close()
		insertDummyTrace(session, "logs-bff", "logs-query")
	}

	fmt.Println("seed complete")
	fmt.Println("email=admin@logs.local password=admin123 apiKey=logs_dev_api_key")
}

func insertDummyTrace(session *gocql.Session, serviceA, serviceB string) {
	// Dummy trace setup
	companyID := "c1"
	traceID := uuid.NewString()
	node1 := uuid.NewString()
	node2 := uuid.NewString()
	now := time.Now()

	// Insert into Cassandra
	_ = session.Query(`INSERT INTO traces_by_company_day(company_id,day,started_at,trace_id,status,root_operation,service_name,http_method,http_path,http_status,duration_ms)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`, companyID, now.Format("2006-01-02"), now, traceID, "OK", "GET /api/users", serviceA, "GET", "/api/users", 200, 45).Exec()

	_ = session.Query(`INSERT INTO nodes_by_trace (trace_id,node_id,company_id,service_name,type,name,status,start_at,end_at,duration_ms,http_method,http_path,http_status)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, traceID, node1, companyID, serviceA, "HTTP", "GET /api/users", "OK", now, now.Add(45*time.Millisecond), 45, "GET", "/api/users", 200).Exec()

	_ = session.Query(`INSERT INTO nodes_by_trace (trace_id,node_id,parent_node_id,company_id,service_name,type,name,status,start_at,end_at,duration_ms,db_system,db_query)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, traceID, node2, node1, companyID, serviceB, "DB", "SELECT users", "OK", now.Add(10*time.Millisecond), now.Add(30*time.Millisecond), 20, "postgres", "SELECT * FROM users").Exec()

	_ = session.Query("INSERT INTO edges_by_trace(trace_id,from_node_id,to_node_id,kind) VALUES(?,?,?,?)", traceID, node1, node2, "PARENT_CHILD").Exec()

	// Insert into OpenSearch
	sc, err := search.NewClient([]string{"http://localhost:9200"})
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = sc.Index(ctx, "nodes", fmt.Sprintf("%s_%s", companyID, node1), map[string]any{"traceId": traceID, "nodeId": node1, "companyId": companyID, "type": "HTTP", "name": "GET /api/users"})
		_ = sc.Index(ctx, "nodes", fmt.Sprintf("%s_%s", companyID, node2), map[string]any{"traceId": traceID, "nodeId": node2, "companyId": companyID, "type": "DB", "name": "SELECT users", "dbQuery": "SELECT * FROM users"})
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
