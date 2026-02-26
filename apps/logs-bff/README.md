# Logs BFF (Backend-For-Frontend)

O Backend-For-Frontend atua como proxy, gerenciador de sessão, hub de integrações e controlador primário das regras de negócio do UI do Logs App.

## ⚙️ Papel na Arquitetura

O Web App (`apps/web`) conversa **exclusivamente** com o `logs-bff`.

Esse serviço toma conta de:

1. Autenticação/Login e expiração (JWT).
2. Fornecimento das métricas, traces, proxy de chamadas de `/search` atuando como ponte (`proxy`/`facade` pattern) com o microserviço `logs-query`.
3. Salvamento e recuperação de `Annotations` em PostgreSQL criadas pelo usuário para o Trace selecionado.
4. Prevenção e Isolamentos: Garante que os Tokens dos usuários logados representem uma `Company` correta e isolam tudo via ID da companhia.

## 🛠️ Seeds

Este serviço expôs na linha de comando e Docker build a feature de semear (seed) o ambiente local. O `cmd/seed/main.go` cria de modo nativo o banco da empresa, um usuário Admin e alimenta o Cassandra / OpenSearch com traços fake de logs `logs-bff` → `logs-query`, facilitando o live-preview.

## 🚀 Como funciona a navegação Front-BFF

Durante um login no React, a sessão devolve o token via Payload do front (`api/login`), então as requests passam a possuir `Authorization: Bearer JWT`. A API BFF decodifica o JWT localizando a empresa original e passa para o query como `X-Company-Id`.
