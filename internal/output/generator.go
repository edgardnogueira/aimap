package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/edgardnogueira/aimap/internal/config"
	"github.com/edgardnogueira/aimap/internal/godoc"
	"github.com/edgardnogueira/aimap/internal/kubedoc"
	"github.com/edgardnogueira/aimap/internal/output/html"
	"github.com/edgardnogueira/aimap/internal/output/markdown"
)

// Generator é responsável por gerar a documentação final
type Generator struct {
    format        string
    outputPath    string
    godocData     *godoc.ProjectDoc
    godocGen      *godoc.Generator
    k8sData       *kubedoc.Resources
    goConfig      config.GolangConfig
}
// NewGenerator cria um novo gerador de documentação
func NewGenerator(format, outputPath string) *Generator {
    return &Generator{
        format:     format,
        outputPath: outputPath,
    }
}

// AddGoDocumentation adiciona documentação Go ao gerador
func (g *Generator) AddGoDocumentation(doc *godoc.ProjectDoc, cfg config.GolangConfig) error {
    g.godocData = doc
    g.goConfig = cfg
    g.godocGen = godoc.NewGenerator(doc, "", cfg)
    return nil
}

// AddKubernetesDocumentation adiciona documentação Kubernetes ao gerador
func (g *Generator) AddKubernetesDocumentation(doc *kubedoc.Resources) error {
    g.k8sData = doc
    return nil
}

// Generate gera a documentação no formato especificado
// Generate gera a documentação no formato especificado
func (g *Generator) Generate() error {
    if err := os.MkdirAll(g.outputPath, 0755); err != nil {
        return fmt.Errorf("erro ao criar diretório de saída: %w", err)
    }

    var content string
    var err error

    switch g.format {
    case "html":
        content, err = g.generateHTML()
    case "markdown":
        content, err = g.generateMarkdown()
    case "json":
        content, err = g.generateJSON()
    case "yaml":
        content, err = g.generateYAML()
    default:
        return fmt.Errorf("formato não suportado: %s", g.format)
    }

    if err != nil {
        return err
    }

    outputFile := filepath.Join(g.outputPath, fmt.Sprintf("documentation.%s", g.getFileExtension()))
    return os.WriteFile(outputFile, []byte(content), 0644)
}
// generateHTML gera documentação em formato HTML
func (g *Generator) generateHTML() (string, error) {
    tmpl := html.NewTemplate()
    
    // Cria o gerador godoc
    godocGen := godoc.NewGenerator(g.godocData, "", g.goConfig)
    mermaidDiagram := godocGen.GenerateMermaidDiagram()
    
    data := html.TemplateData{
        Go:         g.godocData,
        K8s:        g.k8sData,
        GoMermaid:  mermaidDiagram,
        Config: html.ReportConfiguration{
            ReportLevel:    g.goConfig.ReportLevel,
            ReportOptions:  g.goConfig.ReportOptions,
        },
    }

    return tmpl.Generate(data)
}
func (g *Generator) generateMarkdown() (string, error) {
    tmpl := markdown.NewTemplate()
    
    // Gera o diagrama Mermaid
    mermaidDiagram := ""
    if g.godocGen != nil {
        mermaidDiagram = g.godocGen.GenerateMermaidDiagram()
    }
    
    data := markdown.TemplateData{
        Go:        g.godocData,
        K8s:       g.k8sData,
        Config:    g.goConfig,
        GoMermaid: mermaidDiagram,
    }

    return tmpl.Generate(data)
}
// generateJSON gera documentação em formato JSON
func (g *Generator) generateJSON() (string, error) {
    data := struct {
        Go  *godoc.ProjectDoc     `json:"go,omitempty"`
        K8s *kubedoc.Resources    `json:"kubernetes,omitempty"`
    }{
        Go:  g.godocData,
        K8s: g.k8sData,
    }
    
    jsonBytes, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return "", fmt.Errorf("erro ao gerar JSON: %w", err)
    }
    
    return string(jsonBytes), nil
}

// generateYAML gera documentação em formato YAML
func (g *Generator) generateYAML() (string, error) {
    data := struct {
        Go  *godoc.ProjectDoc     `yaml:"go,omitempty"`
        K8s *kubedoc.Resources    `yaml:"kubernetes,omitempty"`
    }{
        Go:  g.godocData,
        K8s: g.k8sData,
    }
    
    yamlBytes, err := yaml.Marshal(data)
    if err != nil {
        return "", fmt.Errorf("erro ao gerar YAML: %w", err)
    }
    
    return string(yamlBytes), nil
}

// getFileExtension retorna a extensão apropriada para o formato
func (g *Generator) getFileExtension() string {
    switch g.format {
    case "html":
        return "html"
    case "markdown":
        return "md"
    case "json":
        return "json"
    case "yaml":
        return "yaml"
    default:
        return "txt"
    }
}

