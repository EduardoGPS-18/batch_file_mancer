# Projeto: File Processor

Este projeto é uma solução de processamento de arquivos com integração a **Kafka** e **PostgreSQL**. Ele foi projetado para ser executado em containers, utilizando gerenciadores de container (como **Docker** ou **Podman**).

## Requisitos

Antes de começar, verifique se você tem os seguintes requisitos:

- **Docker**, **Podman** ou **Outro gerenciador** instalados no seu sistema.
- **Docker Compose** (para Docker) ou uma alternativa compatível com o Podman para orquestrar os containers.

## Inicialização

Para iniciar o projeto, siga estas etapas:

0. (Caso for executar local) Configurar o arquivo `.env` com as variáveis de ambiente necessárias.

```
$ cp .env.example .env
```

1. Executar o docker compose para subir as dependências do projeto:

```bash
$ docker compose up -d
```

2. Executar as migrações do banco de dados:

```bash
$ docker cp ./db/migration.sql file-processor-db:/tmp/migration.sql
$ docker exec -it file-processor-db psql -U postgres -d fileprocessor -f ./tmp/migration.sql
```

## Utilização

Para utilizar o projeto, é disponibilizado uma API Rest para o envio do arquivo.

```bash
$ curl --location 'http://<host (default: localhost)>:<port (default: 8080)/upload/bank-slip/file' \
    --form 'file=@"<path_arquivo>.csv"'
```

Não foi implementado formas de acompanhar o upload e o processamento do arquivo, então é necessário verificar diretamente no banco de dados.

## Testes

### Dependências

1.  Configurar o arquivo `.env.test` a partir do `.env`.

```
$ cp .env.example .env.test
```

2. Para executar os testes de integração e e2e, é necessário algum gerenciador de container
   2.1 Não é necessário subir os containers do projeto, pois os testes sobem os containers necessários (com testcontainers).

### Execução

Foram implementados três níveis de testes:

1.  **Unitários**: Testes unitários para as funções de processamento de arquivos.
    1.1 Para executar, execute o comando:

```bash
$ make test
```

2. **Integração**: Testes de integração para teste das queries.
   1.1 Depende do **banco de dados**, para executar os testes de integração
   1.2 Para executar, execute o comando:

```bash
$ make itest
```

3. **E2E**: Testes de ponta a ponta todo o fluxo principal da aplicação (utilizando testcontainers).
   1.1 Depende do **banco de dados** e do **kafka**, para executar os testes e2e
   1.2 Para executar, execute o comando:

```bash
$ make etest
```
