// cmd/aimap/docker.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/edgardnogueira/aimap/internal/docker"
)

func runDocker(args []string) error {
	dockerCmd := flag.NewFlagSet("docker", flag.ExitOnError)

	// Flags
	projectPath := dockerCmd.String("path", ".", "Caminho para o projeto com Dockerfile/docker-compose")
	outputDir := dockerCmd.String("output", "docs/docker", "Diretório para os arquivos gerados")
	format := dockerCmd.String("format", "plantuml", "Formato de saída (plantuml)")

	if err := dockerCmd.Parse(args); err != nil {
		return err
	}

	// Cria analisador Docker
	analyzer, err := docker.NewAnalyzer(*projectPath)
	if err != nil {
		return fmt.Errorf("erro ao criar analisador Docker: %w", err)
	}

	// Analisa o projeto
	project, err := analyzer.Analyze()
	if err != nil {
		return fmt.Errorf("erro ao analisar configurações Docker: %w", err)
	}

	// Cria o diretório de saída
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída: %w", err)
	}

	// Gera a documentação no formato especificado
	switch *format {
	case "plantuml":
		generator := docker.NewPlantUMLGenerator(project)
		content := generator.Generate()

		outputFile := filepath.Join(*outputDir, "docker.puml")
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