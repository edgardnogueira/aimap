package output

import (
	"github.com/edgardnogueira/aimap/internal/output/html"
	"github.com/edgardnogueira/aimap/internal/output/markdown"
)

// TemplateGenerator define a interface comum para geradores de template
type TemplateGenerator interface {
    Generate(data interface{}) (string, error)
}

// Função auxiliar para criar um novo template baseado no formato
func NewTemplate(format string) TemplateGenerator {
    switch format {
    case "html":
        return html.NewTemplate()
    case "markdown":
        return markdown.NewTemplate()
    default:
        return nil
    }
}

// DocumentationData representa os dados para o template
type DocumentationData struct {
    Go  interface{} `json:"go,omitempty"`
    K8s interface{} `json:"kubernetes,omitempty"`
}

// NewDocumentationData cria uma nova estrutura de dados para documentação
func NewDocumentationData(goData, k8sData interface{}) *DocumentationData {
    return &DocumentationData{
        Go:  goData,
        K8s: k8sData,
    }
}