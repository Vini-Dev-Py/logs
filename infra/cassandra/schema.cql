CREATE KEYSPACE IF NOT EXISTS logs
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TABLE IF NOT EXISTS logs.traces_by_company_day (
  company_id text,
  day text,
  started_at timestamp,
  trace_id text,
  status text,
  root_operation text,
  service_name text,
  http_method text,
  http_path text,
  http_status int,
  duration_ms int,
  PRIMARY KEY ((company_id, day), started_at, trace_id)
) WITH CLUSTERING ORDER BY (started_at DESC);

CREATE TABLE IF NOT EXISTS logs.nodes_by_trace (
  trace_id text,
  node_id text,
  parent_node_id text,
  company_id text,
  service_name text,
  type text,
  name text,
  status text,
  start_at timestamp,
  end_at timestamp,
  duration_ms int,
  http_method text,
  http_path text,
  http_status int,
  db_system text,
  db_query text,
  db_rows int,
  metadata text,
  PRIMARY KEY ((trace_id), node_id)
);

CREATE TABLE IF NOT EXISTS logs.edges_by_trace (
  trace_id text,
  from_node_id text,
  to_node_id text,
  kind text,
  PRIMARY KEY ((trace_id), from_node_id, to_node_id)
);

CREATE TABLE IF NOT EXISTS logs.event_dedup (
  company_id text,
  event_id text,
  received_at timestamp,
  PRIMARY KEY ((company_id), event_id)
) WITH default_time_to_live = 604800;

CREATE TABLE IF NOT EXISTS logs.endpoint_calls_by_company_day (
  company_id text,
  day text,
  service_name text,
  http_method text,
  http_path text,
  calls counter,
  PRIMARY KEY ((company_id, day), service_name, http_method, http_path)
);
