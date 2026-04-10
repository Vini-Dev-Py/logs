CREATE TABLE IF NOT EXISTS roles (
  id UUID PRIMARY KEY,
  company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT,
  is_system_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
  name TEXT PRIMARY KEY,
  description TEXT
);

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_name TEXT NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_name)
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles(id);

-- System Default Permissions
INSERT INTO permissions (name, description) VALUES 
('traces:read', 'Visualizar traces e logs'),
('annotations:write', 'Criar, editar e deletar notas'),
('users:manage', 'Adicionar, remover usuários e gerenciar ocupações')
ON CONFLICT DO NOTHING;
