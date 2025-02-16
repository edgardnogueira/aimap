// cmd/aimap/mysql.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/edgardnogueira/aimap/internal/mysql"
)

func runMySQL(args []string) error {
	mysqlCmd := flag.NewFlagSet("mysql", flag.ExitOnError)

	// Flags
	dsn := mysqlCmd.String("dsn", "", "Data Source Name (formato: user:pass@tcp(host:port)/dbname)")
	dbName := mysqlCmd.String("db", "", "Nome do banco de dados")
	outputDir := mysqlCmd.String("output", "docs/database", "Diretório para os arquivos gerados")
	format := mysqlCmd.String("format", "plantuml", "Formato de saída (plantuml)")

	if err := mysqlCmd.Parse(args); err != nil {
		return err
	}

	if *dsn == "" || *dbName == "" {
		return fmt.Errorf("DSN e nome do banco de dados são obrigatórios")
	}

	// Cria analisador MySQL
	analyzer, err := mysql.NewAnalyzer(*dsn)
	if err != nil {
		return fmt.Errorf("erro ao criar analisador MySQL: %w", err)
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
		generator := mysql.NewPlantUMLGenerator(database)
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