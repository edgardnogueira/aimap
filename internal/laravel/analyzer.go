// internal/laravel/analyzer.go
package laravel

import (
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Analyzer struct {
    projectPath string
}

func NewAnalyzer(projectPath string) (*Analyzer, error) {
    if _, err := os.Stat(projectPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("projeto não encontrado em: %s", projectPath)
    }
    return &Analyzer{projectPath: projectPath}, nil
}

func (a *Analyzer) Analyze() (*Project, error) {
    project := &Project{
        Name: filepath.Base(a.projectPath),
    }

    // Analisa Models
    models, err := a.analyzeModels()
    if err != nil {
        slog.Error("Erro ao analisar models", "error", err)
    }
    project.Models = models

    // Analisa Controllers
    controllers, err := a.analyzeControllers()
    if err != nil {
        slog.Error("Erro ao analisar controllers", "error", err)
    }
    project.Controllers = controllers

    // Analisa Rotas
    routes, err := a.analyzeRoutes()
    if err != nil {
        slog.Error("Erro ao analisar rotas", "error", err)
    }
    project.Routes = routes

    // Analisa Migrations
    migrations, err := a.analyzeMigrations()
    if err != nil {
        slog.Error("Erro ao analisar migrations", "error", err)
    }
    project.Migrations = migrations

    // Analisa Middleware
    middleware, err := a.analyzeMiddleware()
    if err != nil {
        slog.Error("Erro ao analisar middleware", "error", err)
    }
    project.Middleware = middleware

    // Analisa Service Providers
    providers, err := a.analyzeProviders()
    if err != nil {
        slog.Error("Erro ao analisar providers", "error", err)
    }
    project.Providers = providers

    return project, nil
}

func (a *Analyzer) analyzeModels() ([]Model, error) {
    modelsPath := filepath.Join(a.projectPath, "app", "Models")
    files, err := filepath.Glob(filepath.Join(modelsPath, "*.php"))
    if err != nil {
        return nil, err
    }

    var models []Model
    for _, file := range files {
        content, err := ioutil.ReadFile(file)
        if err != nil {
            slog.Error("Erro ao ler arquivo model", "file", file, "error", err)
            continue
        }

        model, err := a.parseModel(content, file)
        if err != nil {
            slog.Error("Erro ao analisar model", "file", file, "error", err)
            continue
        }

        models = append(models, model)
    }

    return models, nil
}

// parseModel analisa um arquivo de model PHP usando regex
func (a *Analyzer) parseModel(content []byte, path string) (Model, error) {
    model := Model{
        Path: path,
        Name: strings.TrimSuffix(filepath.Base(path), ".php"),
    }

    // Encontra nome da tabela
    tableRe := regexp.MustCompile(`protected \$table\s*=\s*['"]([^'"]+)['"]`)
    if matches := tableRe.FindSubmatch(content); len(matches) > 1 {
        model.Table = string(matches[1])
    }

    // Encontra campos fillable
    fillableRe := regexp.MustCompile(`protected \$fillable\s*=\s*\[(.*?)\]`)
    if matches := fillableRe.FindSubmatch(content); len(matches) > 1 {
        fields := strings.Split(string(matches[1]), ",")
        for _, field := range fields {
            field = strings.Trim(field, "' \t\n\r")
            if field != "" {
                model.Fillable = append(model.Fillable, field)
            }
        }
    }

    // Encontra campos hidden
    hiddenRe := regexp.MustCompile(`protected \$hidden\s*=\s*\[(.*?)\]`)
    if matches := hiddenRe.FindSubmatch(content); len(matches) > 1 {
        fields := strings.Split(string(matches[1]), ",")
        for _, field := range fields {
            field = strings.Trim(field, "' \t\n\r")
            if field != "" {
                model.Hidden = append(model.Hidden, field)
            }
        }
    }

    // Encontra casts
    castsRe := regexp.MustCompile(`protected \$casts\s*=\s*\[(.*?)\]`)
    if matches := castsRe.FindSubmatch(content); len(matches) > 1 {
        casts := strings.Split(string(matches[1]), ",")
        for _, cast := range casts {
            cast = strings.Trim(cast, " \t\n\r")
            parts := strings.Split(cast, "=>")
            if len(parts) == 2 {
                field := strings.Trim(parts[0], "' \t\n\r")
                castType := strings.Trim(parts[1], "' \t\n\r")
                if field != "" && castType != "" {
                    model.Casts = append(model.Casts, Cast{
                        Field: field,
                        CastType: castType,
                    })
                }
            }
        }
    }

    // Encontra relacionamentos
    relationTypes := []string{"hasOne", "hasMany", "belongsTo", "belongsToMany"}
    for _, relType := range relationTypes {
        re := regexp.MustCompile(fmt.Sprintf(`public function\s+(\w+)\s*\(\s*\)\s*{\s*return\s+\$this->%s\s*\(\s*([^)]+)\s*\)`, relType))
        matches := re.FindAllSubmatch(content, -1)
        for _, match := range matches {
            if len(match) > 2 {
                method := string(match[1])
                args := strings.Split(string(match[2]), ",")
                relatedModel := strings.Trim(args[0], "' \t\n\r")
                
                rel := Relationship{
                    Type: relType,
                    Method: method,
                    RelatedModel: relatedModel,
                }

                // Processa argumentos adicionais
                for i, arg := range args[1:] {
                    arg = strings.Trim(arg, "' \t\n\r")
                    switch i {
                    case 0:
                        rel.ForeignKey = arg
                    case 1:
                        rel.LocalKey = arg
                    case 2:
                        if relType == "belongsToMany" {
                            rel.PivotTable = arg
                        }
                    }
                }

                model.Relationships = append(model.Relationships, rel)
            }
        }
    }

    return model, nil
}

// Mais funções para análise de controllers, rotas, etc. seguirão o mesmo padrão
// usando expressões regulares para extrair informações dos arquivos PHP