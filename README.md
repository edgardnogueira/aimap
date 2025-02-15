# aimap

aimap é uma ferramenta de linha de comando para gerar documentação automática de projetos Go, Kubernetes e APIs.

## Características

- Documentação automática de código Go
- Documentação de recursos Kubernetes
- Geração de arquivos REST Client (.http) a partir de Swagger/OpenAPI
- Diagramas de classes e relacionamentos
- Múltiplos formatos de saída (Markdown, HTML, JSON, YAML)
- Altamente configurável

## Instalação

```bash
go install github.com/edgardnogueira/aimap/cmd/aimap@latest
```

## Uso Rápido

1. Inicialize um novo projeto:

```bash
aimap init
```

2. Ajuste o arquivo `aimap.yml` gerado conforme necessário:

```yaml
output:
  format: "markdown"
  path: "./docs"

golang:
  enabled: true
  paths:
    - "./cmd"
    - "./internal"
    - "./pkg"
  ...

kubernetes:
  enabled: true
  paths:
    - "./deploy"
  ...

swagger:
  enabled: true
  files:
    - path: "api/swagger.json"
      output: "docs/http-client"
```

3. Gere a documentação:

```bash
aimap generate
```

## Comandos

### Documentação Geral

- `aimap init`: Cria um arquivo de configuração inicial
- `aimap generate`: Gera a documentação
  - `-config`: Caminho para o arquivo de configuração (padrão: aimap.yml)
  - `-format`: Formato de saída (sobrescreve o do arquivo de configuração)
  - `-output`: Caminho de saída (sobrescreve o do arquivo de configuração)
- `aimap version`: Mostra a versão atual

### Swagger/OpenAPI

- `aimap swagger`: Gera arquivos .http para teste de API
  - `-file`: Caminho para o arquivo Swagger/OpenAPI (obrigatório)
  - `-output`: Diretório para os arquivos gerados (padrão: http-client)

Exemplo:

```bash
aimap swagger -file api/swagger.json -output docs/http-tests
```

## Estrutura de Projeto Recomendada

```
seu-projeto/
├── api/
│   └── swagger.json      # Documentação OpenAPI/Swagger
├── docs/
│   ├── http-client/     # Arquivos .http gerados
│   └── markdown/        # Documentação gerada
├── deploy/              # Arquivos Kubernetes
└── aimap.yml           # Configuração do aimap
```

## Configuração

### Formatos de Saída Suportados

- markdown
- html
- json
- yaml

### Opções de Documentação Go

- Níveis de relatório: short, standard, complete
- Opções configuráveis para imports, funções internas, testes e exemplos
- Ignorar arquivos/diretórios específicos

### Opções de Documentação Kubernetes

- Documentação de todos os tipos de recursos
- Geração de diagramas de relacionamento
- Análise de dependências entre recursos

### Opções Swagger/OpenAPI

- Geração de arquivos .http para VS Code REST Client
- Suporte para OpenAPI 2.0 e 3.0
- Agrupamento por tags
- Inclusão de exemplos de requisição

## TODO (Próximos Passos)

### Documentação de Bancos de Dados

- [ ] MySQL: esquemas, tabelas, relacionamentos, procedures
- [ ] PostgreSQL: schemas, functions, views, materialized views
- [ ] SQL Server: schemas, stored procedures, triggers
- [ ] MongoDB: collections, indexes, relations
- [ ] Redis: estruturas de dados, índices

### Frameworks e Tecnologias

- [ ] Laravel: routes, models, controllers, migrations
- [ ] Node.js/Express: routes, middlewares, models
- [ ] Python/Django: models, views, templates
- [ ] Spring Boot: controllers, services, entities
- [ ] Vue.js/React: components, routes, states

### Infraestrutura

- [ ] Docker: Dockerfiles, docker-compose
- [ ] AWS: CloudFormation, recursos AWS
- [ ] Azure: ARM templates, recursos Azure
- [ ] GCP: recursos e configurações

### Melhorias Gerais

- [ ] Geração de diagramas PlantUML
- [ ] Documentação de APIs gRPC
- [ ] Exportação para Confluence/Jira
- [ ] Documentação de testes e cobertura
- [ ] Integração com CI/CD
- [ ] Suporte a webhooks e eventos
- [ ] Análise de segurança e conformidade

## Contribuindo

Contribuições são bem-vindas! Por favor, sinta-se à vontade para submeter pull requests.

## Licença

MIT License - veja o arquivo LICENSE para detalhes.
