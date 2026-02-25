# MVP Status

Este repositório agora contém a estrutura monorepo e uma implementação MVP funcional com:

- logs-bff (auth JWT, traces proxy, annotations CRUD)
- logs-ingest (ingest async com RabbitMQ)
- logs-worker (consumer + persistência Cassandra)
- logs-query (listagem e montagem de grafo)
- web (login, sidebar com animação, traces list, trace viewer React Flow)
- infra (docker-compose, migrations SQL, schema CQL)
