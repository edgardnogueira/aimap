package kubedoc

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/edgardnogueira/aimap/internal/config"
)

// Analyzer é responsável por analisar recursos Kubernetes
type Analyzer struct {
    config config.KubernetesConfig
    parser *Parser
    ignoreRegex []*regexp.Regexp
}

// NewAnalyzer cria um novo analisador de recursos Kubernetes
func NewAnalyzer(cfg config.KubernetesConfig) *Analyzer {
    var regexps []*regexp.Regexp
    for _, pattern := range cfg.Ignores {
        if re, err := regexp.Compile(pattern); err == nil {
            regexps = append(regexps, re)
        } else {
            slog.Warn("Padrão de ignore inválido", "pattern", pattern, "error", err)
        }
    }

    return &Analyzer{
        config: cfg,
        parser: NewParser(),
        ignoreRegex: regexps,
    }
}

// Analyze analisa todos os diretórios configurados
func (a *Analyzer) Analyze() (*Resources, error) {
    for _, path := range a.config.Paths {
        if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }

            // Verifica se deve ignorar o arquivo/diretório
            if a.shouldIgnore(path) {
                if info.IsDir() {
                    return filepath.SkipDir
                }
                return nil
            }

            // Processa apenas arquivos yaml/yml
            if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
                if err := a.parser.ParseFile(path); err != nil {
                    slog.Error("Erro ao analisar arquivo", 
                        "path", path,
                        "error", err)
                }
            }
            return nil
        }); err != nil {
            return nil, fmt.Errorf("erro ao percorrer diretório %s: %w", path, err)
        }
    }

    // Converte os recursos do parser para o formato de saída
    resources := &Resources{
        Resources: make([]Resource, 0),
    }

    for _, node := range a.parser.GetResources() {
        resource := Resource{
            Name:      node.Name,
            Kind:      node.Kind,
            Namespace: node.Namespace,
            Labels:    node.Labels,
            Relations: node.Relations,
        }
        resources.Resources = append(resources.Resources, resource)
    }

    return resources, nil
}

// shouldIgnore verifica se um caminho deve ser ignorado
func (a *Analyzer) shouldIgnore(path string) bool {
    for _, re := range a.ignoreRegex {
        if re.MatchString(path) {
            return true
        }
    }
    return false
}