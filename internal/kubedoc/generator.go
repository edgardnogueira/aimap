// pkg/kubedoc/generator.go
package kubedoc

import (
	"fmt"
	"strings"
)

type Generator struct {
    parser *Parser
}

func NewGenerator(parser *Parser) *Generator {
    return &Generator{parser: parser}
}

func (g *Generator) GenerateMarkdown() string {
    var md strings.Builder

    md.WriteString("# Infraestrutura Kubernetes\n\n")
    
    // Agrupa por tipo
    resourcesByKind := make(map[string][]*ResourceNode)
    for _, resource := range g.parser.resources {
        resourcesByKind[resource.Kind] = append(resourcesByKind[resource.Kind], resource)
    }

    // Gera documentação por tipo
    for kind, resources := range resourcesByKind {
        md.WriteString(fmt.Sprintf("## %s\n\n", kind))
        
        for _, resource := range resources {
            md.WriteString(fmt.Sprintf("### %s\n", resource.Name))
            if resource.Namespace != "" {
                md.WriteString(fmt.Sprintf("Namespace: `%s`\n\n", resource.Namespace))
            }
            
            // Adiciona labels se existirem
            if len(resource.Labels) > 0 {
                md.WriteString("Labels:\n")
                for k, v := range resource.Labels {
                    md.WriteString(fmt.Sprintf("- `%s: %s`\n", k, v))
                }
                md.WriteString("\n")
            }

            // Adiciona relações
            if len(resource.Relations) > 0 {
                md.WriteString("Relations:\n")
                for _, rel := range resource.Relations {
                    md.WriteString(fmt.Sprintf("- %s → %s (%s)\n", 
                        rel.FromName, rel.ToName, rel.Kind))
                }
                md.WriteString("\n")
            }

            md.WriteString("---\n\n")
        }
    }

    return md.String()
}

func (g *Generator) GenerateMermaidDiagram() string {
    var diagram strings.Builder

    diagram.WriteString("```mermaid\ngraph TD\n")
    
    // Adiciona nós
    for _, resource := range g.parser.resources {
        id := strings.ToLower(resource.Name)
        diagram.WriteString(fmt.Sprintf("    %s[%s: %s]\n", 
            id, resource.Kind, resource.Name))
        
        // Adiciona estilo baseado no tipo
        style := g.getNodeStyle(resource.Kind)
        diagram.WriteString(fmt.Sprintf("    style %s %s\n", id, style))
    }

    // Adiciona relações
    for _, resource := range g.parser.resources {
        for _, rel := range resource.Relations {
            fromID := strings.ToLower(rel.FromName)
            toID := strings.ToLower(rel.ToName)
            diagram.WriteString(fmt.Sprintf("    %s --> %s\n", fromID, toID))
        }
    }

    diagram.WriteString("```")
    return diagram.String()
}

func (g *Generator) getNodeStyle(kind string) string {
    switch kind {
    case "Deployment":
        return "fill:#afd,stroke:#3a3"
    case "StatefulSet":
        return "fill:#fad,stroke:#a33"
    case "Service":
        return "fill:#ddf,stroke:#33a"
    case "Ingress":
        return "fill:#ffa,stroke:#aa3"
    default:
        return "fill:#fff,stroke:#999"
    }
}