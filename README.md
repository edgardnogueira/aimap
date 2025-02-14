# aimap

aimap é uma ferramenta de linha de comando para gerar documentação automática de projetos Go e Kubernetes.

## Características

- Documentação automática de código Go
- Documentação de recursos Kubernetes
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
```

3. Gere a documentação:

```bash
aimap generate
```

## Comandos

- `aimap init`: Cria um arquivo de configuração inicial
- `aimap generate`: Gera a documentação
  - `-config`: Caminho para o arquivo de configuração (padrão: aimap.yml)
  - `-format`: Formato de saída (sobrescreve o do arquivo de configuração)
  - `-output`: Caminho de saída (sobrescreve o do arquivo de configuração)
- `aimap version`: Mostra a versão atual

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

## Contribuindo

Contribuições são bem-vindas! Por favor, sinta-se à vontade para submeter pull requests.

## Licença

MIT License - veja o arquivo LICENSE para detalhes.
