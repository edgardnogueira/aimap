// cmd/aimap/postgres.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/edgardnogueira/aimap/internal/postgres"
)

func runPostgres(args []string) error {
	postgresCmd := flag.NewFlagSet("postgres", flag.ExitOnError)

	// Flags
	connStr := postgresCmd.String("conn", "", "String de conexão (formato: postgres://user:pass@host:port/dbname)")
	dbName := postgresCmd.String("db", "", "Nome do banco de dados")
	outputDir := postgresCmd.String("output", "docs/database", "Diretório para os arquivos gerados")
	format := postgresCmd.String("format", "plantuml", "Formato de saída (plantuml)")

	if err := postgresCmd.Parse(args); err != nil {
		return err
	}

	if *connStr == "" || *dbName == "" {
		return fmt.Errorf("string de conexão e nome do banco de dados são obrigatórios")
	}

	// Cria analisador PostgreSQL
	analyzer, err := postgres.NewAnalyzer(*connStr)
	if err != nil {
		return fmt.Errorf("erro ao criar analisador PostgreSQL: %w", err)
	}
	defer analyzer.Close()

	// Analisa o banco de dados
	database, err := analyzer.Analyze(*dbName)
	if err != nil {
		return fmt.Errorf("erro ao analisar banco de dados: %w", err)
	}

	// Cria o diretório de saída
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída: %w", err)
	}

	// Gera a documentação no formato especificado
	switch *format {
	case "plantuml":
		generator := postgres.NewPlantUMLGenerator(database)
		content := generator.Generate()

		outputFile := filepath.Join(*outputDir, "database.puml")
		if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("erro ao escrever arquivo PlantUML: %w", err)
		}

		slog.Info("Documentação PlantUML gerada com sucesso", 
			"file", outputFile)

	default:
		return fmt.Errorf("formato não suportado: %s", *format)
	}

	return nil
}