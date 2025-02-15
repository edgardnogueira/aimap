// cmd/aimap/swagger.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/edgardnogueira/aimap/internal/swagger"
)

func runSwagger(args []string) error {
	// Subcomando para swagger
	swaggerCmd := flag.NewFlagSet("swagger", flag.ExitOnError)
	
	// Flags
	swaggerFile := swaggerCmd.String("file", "", "Caminho para o arquivo Swagger/OpenAPI")
	outputDir := swaggerCmd.String("output", "http-client", "Diretório para os arquivos .http gerados")
	
	if err := swaggerCmd.Parse(args); err != nil {
		return err
	}

	if *swaggerFile == "" {
		return fmt.Errorf("arquivo swagger não especificado")
	}

	// Verifica se o arquivo existe
	if _, err := os.Stat(*swaggerFile); os.IsNotExist(err) {
		return fmt.Errorf("arquivo swagger não encontrado: %s", *swaggerFile)
	}

	// Parse do arquivo Swagger
	doc, err := swagger.Parse(*swaggerFile)
	if err != nil {
		return fmt.Errorf("erro ao fazer parse do swagger: %w", err)
	}

	// Cria o diretório de saída se não existir
	outputPath := *outputDir
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(".", outputPath)
	}

	// Gera os arquivos .http
	if err := doc.GenerateHTTPFiles(outputPath); err != nil {
		return fmt.Errorf("erro ao gerar arquivos .http: %w", err)
	}

	slog.Info("Arquivos .http gerados com sucesso", "output_dir", outputPath)
	return nil
}