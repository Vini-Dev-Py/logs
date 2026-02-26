# Logs Web App

A interface visual Interativa de Grafos criada com **React Flow**, Taildwind CSS, TypeScript e React Router DOM.

## ⚙️ Papel na Arquitetura

Trazendo o front para perto da observabilidade, essa aplicação retira a dependência de análises via texto no terminal através de uma UX impecável para visualização hierárquica e lateral.

Principais telas disponíveis:

1. **Login:** Acessado usando email e senha encriptada do Owner.
2. **Dashboard de Traces (`TracesPage`):** Lista todos as "Execuções" do sistema que passaram por `logs-ingest`. Mostra contadores, tempos, root operations entre outros. Uma caixa de pesquisas executa consultas ativas por Nodes em traces.
3. **Graph Trace Viewer (`TraceViewerPage`):** Para analisar "O que aconteceu" do início ao fim. O fluxo das operações em caixas (Nodes), dependências e chamadas filhas conectadas por arestas. Suporta adição de "Anotações" para marcações da equipe e contem Realce Visual (Highlight) em azul quando procuramos (Ctrl+F nativo) os dados.

## 🚀 Como acessar local

Após usar o comando `make up` e injetar dados fakes através de `make seed`, abra no navegador:
`http://localhost:5173`

**Credenciais default da empresa gerada pelo Seed:**

- **Email:** `admin@logs.local`
- **Senha:** `admin123`

_(Qualquer ação tomada na UI baterá no `logs-bff` exposto publicamente no Traefik na sub-rota `/api`)_
