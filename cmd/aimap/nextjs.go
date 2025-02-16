// cmd/aimap/nextjs.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/edgardnogueira/aimap/internal/nextjs"
)

func runNextjs(args []string) error {
	nextjsCmd := flag.NewFlagSet("nextjs", flag.ExitOnError)

	// Flags
	projectPath := nextjsCmd.String("path", ".", "Caminho para o projeto Next.js")
	outputDir := nextjsCmd.String("output", "docs/nextjs", "Diretório para os arquivos gerados")
	format := nextjsCmd.String("format", "plantuml", "Formato de saída (plantuml)")

	if err := nextjsCmd.Parse(args); err != nil {
		return err
	}

	// Cria analisador Next.js
	analyzer, err := nextjs.NewAnalyzer(*projectPath)
	if err != nil {
		return fmt.Errorf("erro ao criar analisador Next.js: %w", err)
	}

	// Analisa o projeto
	project, err := analyzer.Analyze()
	if err != nil {
		return fmt.Errorf("erro ao analisar projeto Next.js: %w", err)
	}

	// Cria o diretório de saída
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída: %w", err)
	}

	// Gera a documentação no formato especificado
	switch *format {
	case "plantuml":
		generator := nextjs.NewPlantUMLGenerator(project)
		content := generator.Generate()

		outputFile := filepath.Join(*outputDir, "nextjs.puml")
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