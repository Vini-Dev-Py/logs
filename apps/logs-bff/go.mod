module logs-bff

go 1.23

toolchain go1.24.3

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/gocql/gocql v1.7.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.6.0
	golang.org/x/crypto v0.24.0
	shared-search v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang/snappy v0.0.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/opensearch-project/opensearch-go/v2 v2.3.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
)

replace shared-search => ../../packages/shared-search
