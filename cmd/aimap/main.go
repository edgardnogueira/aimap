// cmd/superdoc/main.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/edgardnogueira/aimap/internal/config"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Configurar logging estruturado
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Subcomandos
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)

	// Flags para o comando generate
	configFile := generateCmd.String("config", "superdoc.yml", "Caminho para o arquivo de configuração")
	outputFormat := generateCmd.String("format", "", "Formato de saída (sobrescreve o do arquivo de configuração)")
	outputPath := generateCmd.String("output", "", "Caminho de saída (sobrescreve o do arquivo de configuração)")

	// Verificar argumentos
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "nextjs":
		if err := runNextjs(os.Args[2:]); err != nil {
			slog.Error("Erro ao gerar documentação Next.js", "error", err)
			os.Exit(1)
		}
	case "docker":
		if err := runDocker(os.Args[2:]); err != nil {
			slog.Error("Erro ao gerar documentação Docker", "error", err)
			os.Exit(1)
		}
	case "laravel":
		if err := runLaravel(os.Args[2:]); err != nil {
			slog.Error("Erro ao gerar documentação Laravel", "error", err)
			os.Exit(1)
		}
	case "mysql":
		if err := runMySQL(os.Args[2:]); err != nil {
			slog.Error("Erro ao gerar documentação MySQL", "error", err)
			os.Exit(1)
		}
	case "postgres":
		if err := runPostgres(os.Args[2:]); err != nil {
			slog.Error("Erro ao gerar documentação PostgreSQL", "error", err)
			os.Exit(1)
		}
	case "swagger":
		if err := runSwagger(os.Args[2:]); err != nil {
			slog.Error("Erro ao executar comando swagger", "error", err)
			os.Exit(1)
		}
	case "init":
		initCmd.Parse(os.Args[2:])
		if err := runInit(); err != nil {
			slog.Error("Erro ao inicializar projeto", "error", err)
			os.Exit(1)
		}

	case "generate":
		generateCmd.Parse(os.Args[2:])
		if err := runGenerate(*configFile, *outputFormat, *outputPath); err != nil {
			slog.Error("Erro ao gerar documentação", "error", err)
			os.Exit(1)
		}

	case "version":
		fmt.Printf("superdoc version %s (built at %s)\n", Version, BuildTime)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`superdoc - Gerador de Documentação para Go e Kubernetes

Uso:
  superdoc <comando> [argumentos]

Comandos:
  init      Inicializa um novo projeto com arquivo de configuração
  generate  Gera a documentação baseada na configuração
  version   Mostra a versão do superdoc

Execute 'superdoc <comando> -h' para mais informações sobre um comando específico.`)
}

func runInit() error {
	configPath := "superdoc.yml"
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("arquivo de configuração já existe: %s", configPath)
	}

	// Template do arquivo de configuração
	template := `# Configuração do SuperDoc
output:
  format: "markdown" # Pode ser: html, markdown, json, yaml
  path: "./docs"     # Diretório onde a documentação será gerada

golang:
  enabled: true
  report_level: "standard"  # Pode ser: short, standard, complete
  report_options:
    show_imports: true
    show_internal_funcs: true
    show_tests: false
    show_examples: true
  paths:
    - "./cmd"
    - "./internal"
    - "./pkg"
  ignores:
    - ".*_test\\.go$"
    - "vendor/.*"
    - "node_modules/.*"

kubernetes:
  enabled: true
  paths:
    - "./deploy"
  ignores:
    - ".*\\.bak$"
    - ".*\\.tmp$"
`

	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("erro ao criar arquivo de configuração: %w", err)
	}

	fmt.Printf("Arquivo de configuração criado em %s\n", configPath)
	return nil
}

func runGenerate(configFile, outputFormat, outputPath string) error {
	// Carregar configuração
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	// Sobrescrever configurações se fornecidas via linha de comando
	if outputFormat != "" {
		cfg.Output.Format = outputFormat
	}
	if outputPath != "" {
		cfg.Output.Path = outputPath
	}

	// Garantir que o diretório de saída existe
	if err := os.MkdirAll(cfg.Output.Path, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída: %w", err)
	}

	// Resto da lógica de geração...
	// [Use seu código existente aqui]

	return nil
}