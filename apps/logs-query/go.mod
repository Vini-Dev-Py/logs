module logs-query

go 1.22

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/gocql/gocql v1.7.0
	shared-search v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang/snappy v0.0.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/opensearch-project/opensearch-go/v2 v2.3.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
)

replace shared-search => ../../packages/shared-search
