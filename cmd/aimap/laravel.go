// cmd/aimap/laravel.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/edgardnogueira/aimap/internal/laravel"
)

func runLaravel(args []string) error {
	laravelCmd := flag.NewFlagSet("laravel", flag.ExitOnError)

	// Flags
	projectPath := laravelCmd.String("path", ".", "Caminho para o projeto Laravel")
	outputDir := laravelCmd.String("output", "docs/laravel", "Diretório para os arquivos gerados")
	format := laravelCmd.String("format", "plantuml", "Formato de saída (plantuml)")

	if err := laravelCmd.Parse(args); err != nil {
		return err
	}

	// Cria analisador Laravel
	analyzer, err := laravel.NewAnalyzer(*projectPath)
	if err != nil {
		return fmt.Errorf("erro ao criar analisador Laravel: %w", err)
	}

	// Analisa o projeto
	project, err := analyzer.Analyze()
	if err != nil {
		return fmt.Errorf("erro ao analisar projeto Laravel: %w", err)
	}

	// Cria o diretório de saída
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída: %w", err)
	}

	// Gera a documentação no formato especificado
	switch *format {
	case "plantuml":
		generator := laravel.NewPlantUMLGenerator(project)
		content := generator.Generate()

		outputFile := filepath.Join(*outputDir, "laravel.puml")
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