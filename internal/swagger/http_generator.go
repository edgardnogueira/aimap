// internal/swagger/http_generator.go
package swagger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func generateHTTPFile(filename, tag string, endpoints []HTTPEndpoint) error {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("### %s Endpoints\n\n", tag))
	content.WriteString("@baseUrl = {{$dotenv BASE_URL}}\n")
	content.WriteString("@authToken = {{$dotenv AUTH_TOKEN}}\n\n")

	for _, endpoint := range endpoints {
		// Adiciona comentário com descrição
		if endpoint.Operation.Description != "" {
			content.WriteString(fmt.Sprintf("### %s\n", endpoint.Operation.Description))
		} else if endpoint.Operation.Summary != "" {
			content.WriteString(fmt.Sprintf("### %s\n", endpoint.Operation.Summary))
		}

		// Início da requisição
		content.WriteString(fmt.Sprintf("# @name %s\n", getOperationName(endpoint)))
		content.WriteString(fmt.Sprintf("%s {{baseUrl}}%s", endpoint.Method, endpoint.Path))

		// Adiciona query parameters
		queryParams := getQueryParams(endpoint.Operation.Parameters)
		if len(queryParams) > 0 {
			content.WriteString("?" + strings.Join(queryParams, "&"))
		}
		content.WriteString("\n")

		// Adiciona headers comuns
		content.WriteString("Content-Type: application/json\n")
		if needsAuth(endpoint.Operation) {
			content.WriteString("Authorization: Bearer {{authToken}}\n")
		}

		// Adiciona body se necessário
		if body := getRequestBody(endpoint.Operation); body != "" {
			content.WriteString("\n")
			content.WriteString(body)
		}

		content.WriteString("\n###\n\n")
	}

	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func getOperationName(endpoint HTTPEndpoint) string {
	if endpoint.Operation.OperationID != "" {
		return endpoint.Operation.OperationID
	}
	// Gera um nome baseado no método e caminho
	path := strings.ReplaceAll(endpoint.Path, "/", "_")
	path = strings.Trim(path, "_")
	return fmt.Sprintf("%s_%s", strings.ToLower(endpoint.Method), path)
}

func getQueryParams(params []Parameter) []string {
	var queryParams []string
	for _, param := range params {
		if param.In == "query" {
			// Usa um valor de exemplo ou placeholder
			value := "value"
			if param.Schema != nil && param.Schema.Example != nil {
				value = fmt.Sprintf("%v", param.Schema.Example)
			}
			queryParams = append(queryParams, fmt.Sprintf("%s={{%s}}", param.Name, value))
		}
	}
	return queryParams
}

func needsAuth(op Operation) bool {
	// Verifica se tem algum security scheme definido
	// Isso é uma simplificação - você pode expandir conforme necessário
	return true
}

func getRequestBody(op Operation) string {
	if op.RequestBody == nil {
		return ""
	}

	// Procura por content-type JSON
	for contentType, mediaType := range op.RequestBody.Content {
		if strings.Contains(contentType, "json") {
			// Se tiver exemplo, usa ele
			if mediaType.Example != nil {
				bytes, err := json.MarshalIndent(mediaType.Example, "", "  ")
				if err == nil {
					return string(bytes)
				}
			}

			// Se não tiver exemplo, gera um template baseado no schema
			return generateSchemaExample(mediaType.Schema)
		}
	}

	return ""
}

func generateSchemaExample(schema SchemaType) string {
	example := generateExample(schema)
	bytes, err := json.MarshalIndent(example, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func generateExample(schema SchemaType) interface{} {
	if schema.Example != nil {
		return schema.Example
	}

	switch schema.Type {
	case "object":
		obj := make(map[string]interface{})
		for name, prop := range schema.Properties {
			obj[name] = generateExample(prop)
		}
		return obj

	case "array":
		if schema.Items != nil {
			return []interface{}{generateExample(*schema.Items)}
		}
		return []interface{}{}

	case "string":
		switch schema.Format {
		case "date-time":
			return "2024-02-14T12:00:00Z"
		case "date":
			return "2024-02-14"
		case "email":
			return "user@example.com"
		default:
			return "string"
		}

	case "integer", "number":
		return 0

	case "boolean":
		return false

	default:
		return nil
	}
}