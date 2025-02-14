package kubedoc

import (
	"fmt"
	"io/ioutil"
	"log/slog"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// KubeResource representa uma estrutura genérica para recursos Kubernetes
type KubeResource struct {
    ApiVersion string `yaml:"apiVersion"`
    Kind       string `yaml:"kind"`
    Metadata   struct {
        Name      string `yaml:"name"`
        Namespace string `yaml:"namespace"`
    } `yaml:"metadata"`
    Spec map[string]interface{} `yaml:"spec"`
}

// DocGenerator é responsável por gerar a documentação
type DocGenerator struct {
    InfraPath string
    Resources []KubeResource
}

// NewDocGenerator cria uma nova instância do gerador de documentação
func NewDocGenerator(infraPath string) *DocGenerator {
    return &DocGenerator{
        InfraPath: infraPath,
        Resources: make([]KubeResource, 0),
    }
}
// ReadInfraFiles lê todos os arquivos yaml do diretório
func (g *DocGenerator) ReadInfraFiles() error {
    files, err := ioutil.ReadDir(g.InfraPath)
    if err != nil {
        return err
    }
    slog.Info("Lendo arquivos do diretório", "path", g.InfraPath)

    for _, file := range files {
        slog.Info("Processando arquivo", "filename", file.Name())
        if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
            content, err := ioutil.ReadFile(filepath.Join(g.InfraPath, file.Name()))
            if err != nil {
                return err
            }

            // Dividir o conteúdo em documentos YAML separados
            documents := strings.Split(string(content), "---")
            
            for i, doc := range documents {
                if strings.TrimSpace(doc) == "" {
                    continue
                }
                
                var resource KubeResource
                if err := yaml.Unmarshal([]byte(doc), &resource); err != nil {
                    slog.Error("Erro ao fazer unmarshal do documento", 
                        "filename", file.Name(),
                        "doc_index", i,
                        "error", err)
                    continue
                }
                
                slog.Info("Recurso encontrado", 
                    "kind", resource.Kind,
                    "name", resource.Metadata.Name,
                    "namespace", resource.Metadata.Namespace)

                g.Resources = append(g.Resources, resource)
            }
        }
    }

    slog.Info("Total de recursos encontrados", "count", len(g.Resources))
    return nil
}
func (g *DocGenerator) GenerateMarkdown() string {
    var md strings.Builder

    md.WriteString("# Infraestrutura Kubernetes\n\n")
    md.WriteString("## Recursos\n\n")

    // Agrupa recursos por tipo
    resourcesByKind := make(map[string][]KubeResource)
    for _, resource := range g.Resources {
        resourcesByKind[resource.Kind] = append(resourcesByKind[resource.Kind], resource)
    }

    // Documenta cada tipo de recurso
    for kind, resources := range resourcesByKind {
        md.WriteString(fmt.Sprintf("### %s\n\n", kind))
        
        for _, resource := range resources {
            md.WriteString(fmt.Sprintf("#### %s\n", resource.Metadata.Name))
            if resource.Metadata.Namespace != "" {
                md.WriteString(fmt.Sprintf("Namespace: `%s`\n\n", resource.Metadata.Namespace))
            }

            // Adiciona detalhes específicos baseado no tipo
            switch kind {
            case "Deployment", "StatefulSet":
                if replicas, ok := resource.Spec["replicas"].(int); ok {
                    md.WriteString(fmt.Sprintf("Replicas: %d\n", replicas))
                }
                if template, ok := resource.Spec["template"].(map[interface{}]interface{}); ok {
                    if spec, ok := template["spec"].(map[interface{}]interface{}); ok {
                        if containers, ok := spec["containers"].([]interface{}); ok {
                            md.WriteString("\nContainers:\n")
                            for _, c := range containers {
                                container := c.(map[interface{}]interface{})
                                md.WriteString(fmt.Sprintf("- Name: %s\n", container["name"]))
                                md.WriteString(fmt.Sprintf("  Image: %s\n", container["image"]))
                            }
                        }
                    }
                }
            case "Service":
                if serviceType, ok := resource.Spec["type"].(string); ok {
                    md.WriteString(fmt.Sprintf("Type: %s\n", serviceType))
                }
                if ports, ok := resource.Spec["ports"].([]interface{}); ok {
                    md.WriteString("\nPorts:\n")
                    for _, p := range ports {
                        port := p.(map[interface{}]interface{})
                        md.WriteString(fmt.Sprintf("- %v -> %v\n", 
                            port["port"], 
                            port["targetPort"]))
                    }
                }
            case "Ingress":
                if rules, ok := resource.Spec["rules"].([]interface{}); ok {
                    md.WriteString("\nRules:\n")
                    for _, r := range rules {
                        rule := r.(map[interface{}]interface{})
                        if host, ok := rule["host"].(string); ok {
                            md.WriteString(fmt.Sprintf("- Host: %s\n", host))
                        }
                    }
                }
            }
            
            md.WriteString("\n---\n\n")
        }
    }

    return md.String()
}
func (g *DocGenerator) GenerateMermaidDiagram() string {
    var diagram strings.Builder

    diagram.WriteString("```mermaid\ngraph TD\n")
    
    // Mapa para armazenar IDs únicos
    usedIDs := make(map[string]bool)

    // Função helper para gerar ID único
    generateUniqueID := func(base string) string {
        id := base
        counter := 1
        for usedIDs[id] {
            id = fmt.Sprintf("%s_%d", base, counter)
            counter++
        }
        usedIDs[id] = true
        return id
    }

    // Primeiro, cria todos os nós
    for _, resource := range g.Resources {
        if resource.Metadata.Name == "" {
            continue // Pula recursos sem nome
        }

        id := generateUniqueID(strings.ToLower(resource.Metadata.Name))
        
        // Define estilos baseados no tipo de recurso
        var style string
        switch resource.Kind {
        case "Deployment":
            style = "style " + id + " fill:#afd,stroke:#3a3"
        case "StatefulSet":
            style = "style " + id + " fill:#fad,stroke:#a33"
        case "Service":
            style = "style " + id + " fill:#ddf,stroke:#33a"
        case "Ingress":
            style = "style " + id + " fill:#ffa,stroke:#aa3"
        case "Namespace":
            style = "style " + id + " fill:#eee,stroke:#666"
        default:
            style = "style " + id + " fill:#fff,stroke:#999"
        }

        // Adiciona o nó
        diagram.WriteString(fmt.Sprintf("    %s[%s: %s]\n", 
            id,
            resource.Kind,
            resource.Metadata.Name))
        diagram.WriteString("    " + style + "\n")
    }

    // Depois, adiciona as conexões
    for _, resource := range g.Resources {
        sourceID := strings.ToLower(resource.Metadata.Name)
        
        // Conecta serviços aos seus seletores
        if resource.Kind == "Service" {
            if selectors, ok := resource.Spec["selector"].(map[interface{}]interface{}); ok {
                for _, res := range g.Resources {
                    if res.Kind == "Deployment" || res.Kind == "StatefulSet" {
                        if template, ok := res.Spec["template"].(map[interface{}]interface{}); ok {
                            if metadata, ok := template["metadata"].(map[interface{}]interface{}); ok {
                                if labels, ok := metadata["labels"].(map[interface{}]interface{}); ok {
                                    matches := true
                                    for k, v := range selectors {
                                        if labels[k] != v {
                                            matches = false
                                            break
                                        }
                                    }
                                    if matches {
                                        targetID := strings.ToLower(res.Metadata.Name)
                                        diagram.WriteString(fmt.Sprintf("    %s --> %s\n", sourceID, targetID))
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }

        // Conecta Ingress aos Services
        if resource.Kind == "Ingress" {
            if rules, ok := resource.Spec["rules"].([]interface{}); ok {
                for _, r := range rules {
                    rule := r.(map[interface{}]interface{})
                    if http, ok := rule["http"].(map[interface{}]interface{}); ok {
                        if paths, ok := http["paths"].([]interface{}); ok {
                            for _, p := range paths {
                                path := p.(map[interface{}]interface{})
                                if backend, ok := path["backend"].(map[interface{}]interface{}); ok {
                                    if service, ok := backend["service"].(map[interface{}]interface{}); ok {
                                        if serviceName, ok := service["name"].(string); ok {
                                            targetID := strings.ToLower(serviceName)
                                            diagram.WriteString(fmt.Sprintf("    %s --> %s\n", sourceID, targetID))
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    diagram.WriteString("```")
    return diagram.String()
}