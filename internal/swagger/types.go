// internal/swagger/types.go
package swagger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SwaggerDoc representa a estrutura do documento Swagger/OpenAPI
type SwaggerDoc struct {
	Swagger     string                 `json:"swagger"`
	OpenAPI     string                 `json:"openapi"`
	Info        Info                   `json:"info"`
	Host        string                 `json:"host,omitempty"`
	BasePath    string                 `json:"basePath,omitempty"`
	Servers     []Server               `json:"servers,omitempty"`
	Paths       map[string]PathItem    `json:"paths"`
	Components  *Components            `json:"components,omitempty"`
	Definitions map[string]SchemaType  `json:"definitions,omitempty"` // Swagger 2.0
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type PathItem map[string]Operation // key é o método HTTP (get, post, etc)

type Operation struct {
	Summary     string               `json:"summary,omitempty"`
	Description string               `json:"description,omitempty"`
	OperationID string               `json:"operationId,omitempty"`
	Parameters  []Parameter          `json:"parameters,omitempty"`
	RequestBody *RequestBody         `json:"requestBody,omitempty"`
	Responses   map[string]Response  `json:"responses"`
	Tags        []string             `json:"tags,omitempty"`
}

type Parameter struct {
	Name        string       `json:"name"`
	In          string       `json:"in"` // query, path, header, cookie
	Description string       `json:"description,omitempty"`
	Required    bool         `json:"required,omitempty"`
	Schema      *SchemaType  `json:"schema,omitempty"`
}

type RequestBody struct {
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required,omitempty"`
	Content     map[string]MediaType   `json:"content"`
}

type Response struct {
	Description string                 `json:"description"`
	Content     map[string]MediaType   `json:"content,omitempty"`
}

type MediaType struct {
	Schema  SchemaType  `json:"schema"`
	Example interface{} `json:"example,omitempty"`
}

type SchemaType struct {
	Type        string                 `json:"type,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Properties  map[string]SchemaType  `json:"properties,omitempty"`
	Items       *SchemaType           `json:"items,omitempty"`
	Ref         string                 `json:"$ref,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
}

type Components struct {
	Schemas    map[string]SchemaType  `json:"schemas,omitempty"`
	Parameters map[string]Parameter   `json:"parameters,omitempty"`
	Responses  map[string]Response    `json:"responses,omitempty"`
}

// Parse lê e analisa um arquivo Swagger/OpenAPI
func Parse(filePath string) (*SwaggerDoc, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo swagger: %w", err)
	}

	var doc SwaggerDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do swagger: %w", err)
	}

	return &doc, nil
}

// GenerateHTTPFiles gera arquivos .http para cada endpoint
func (doc *SwaggerDoc) GenerateHTTPFiles(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída: %w", err)
	}

	// Determina o host base
	baseURL := doc.getBaseURL()

	// Agrupa endpoints por tags
	endpointsByTag := make(map[string][]HTTPEndpoint)
	for path, pathItem := range doc.Paths {
		for method, op := range pathItem {
			endpoint := HTTPEndpoint{
				Method:      strings.ToUpper(method),
				Path:        path,
				Operation:   op,
				BaseURL:     baseURL,
			}

			// Se não houver tags, usa "default"
			tags := op.Tags
			if len(tags) == 0 {
				tags = []string{"default"}
			}

			// Adiciona o endpoint a cada tag
			for _, tag := range tags {
				endpointsByTag[tag] = append(endpointsByTag[tag], endpoint)
			}
		}
	}

	// Gera um arquivo por tag
	for tag, endpoints := range endpointsByTag {
		filename := filepath.Join(outputDir, fmt.Sprintf("%s.http", sanitizeFilename(tag)))
		if err := generateHTTPFile(filename, tag, endpoints); err != nil {
			return fmt.Errorf("erro ao gerar arquivo %s: %w", filename, err)
		}
	}

	return nil
}

// getBaseURL determina a URL base para as requisições
func (doc *SwaggerDoc) getBaseURL() string {
	// OpenAPI 3.0
	if len(doc.Servers) > 0 {
		return doc.Servers[0].URL
	}

	// Swagger 2.0
	if doc.Host != "" {
		scheme := "https"
		if doc.BasePath != "" {
			return fmt.Sprintf("%s://%s%s", scheme, doc.Host, doc.BasePath)
		}
		return fmt.Sprintf("%s://%s", scheme, doc.Host)
	}

	return "http://localhost:8080"
}

// HTTPEndpoint representa um endpoint para o arquivo .http
type HTTPEndpoint struct {
	Method    string
	Path      string
	Operation Operation
	BaseURL   string
}

// sanitizeFilename limpa o nome do arquivo
func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}