// internal/laravel/methods.go
package laravel

import (
	"io/ioutil"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
)

// Implementação dos métodos de análise

func (a *Analyzer) analyzeControllers() ([]Controller, error) {
	controllersPath := filepath.Join(a.projectPath, "app", "Http", "Controllers")
	files, err := filepath.Glob(filepath.Join(controllersPath, "*.php"))
	if err != nil {
		return nil, err
	}

	var controllers []Controller
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			slog.Error("Erro ao ler arquivo controller", "file", file, "error", err)
			continue
		}

		controller := Controller{
			Name: strings.TrimSuffix(filepath.Base(file), ".php"),
			Path: file,
		}

		// Analisa métodos
		methodRe := regexp.MustCompile(`public function\s+(\w+)\s*\((.*?)\)(?:\s*:\s*([^\s{]+))?`)
		matches := methodRe.FindAllSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				method := Method{
					Name: string(match[1]),
				}
				// Adiciona parâmetros se houver
				if len(match) > 2 && len(match[2]) > 0 {
					params := strings.Split(string(match[2]), ",")
					for _, param := range params {
						param = strings.TrimSpace(param)
						if param != "" {
							method.Parameters = append(method.Parameters, param)
						}
					}
				}
				controller.Methods = append(controller.Methods, method)
			}
		}

		controllers = append(controllers, controller)
	}

	return controllers, nil
}

func (a *Analyzer) analyzeRoutes() ([]Route, error) {
	routesPath := filepath.Join(a.projectPath, "routes", "web.php")
	content, err := ioutil.ReadFile(routesPath)
	if err != nil {
		return nil, err
	}

	var routes []Route
	// Analisa definições de rotas
	routeRe := regexp.MustCompile(`Route::(get|post|put|patch|delete)\s*\(\s*['"]([^'"]+)['"],\s*(?:\[([^\]]+)\]|['"]([^'"]+)['"])\s*\)`)
	matches := routeRe.FindAllSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			route := Route{
				Method: strings.ToUpper(string(match[1])),
				URI:    string(match[2]),
			}

			// Extrai ação do controlador
			if len(match[3]) > 0 {
				route.Action = string(match[3])
			}

			routes = append(routes, route)
		}
	}

	return routes, nil
}

func (a *Analyzer) analyzeMigrations() ([]Migration, error) {
	migrationsPath := filepath.Join(a.projectPath, "database", "migrations")
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.php"))
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			slog.Error("Erro ao ler arquivo migration", "file", file, "error", err)
			continue
		}

		migration := Migration{
			Name: strings.TrimSuffix(filepath.Base(file), ".php"),
			Path: file,
		}

		// Analisa colunas da tabela
		columnRe := regexp.MustCompile(`\$table->(\w+)\(\s*['"]([^'"]+)['"](?:,\s*([^)]+))?\)(?:->([^;]+))?`)
		matches := columnRe.FindAllSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 2 {
				column := Column{
					Name: string(match[2]),
					Type: string(match[1]),
				}
				migration.Columns = append(migration.Columns, column)
			}
		}

		migrations = append(migrations, migration)
	}

	return migrations, nil
}

func (a *Analyzer) analyzeMiddleware() ([]Middleware, error) {
	middlewarePath := filepath.Join(a.projectPath, "app", "Http", "Middleware")
	files, err := filepath.Glob(filepath.Join(middlewarePath, "*.php"))
	if err != nil {
		return nil, err
	}

	var middleware []Middleware
	for _, file := range files {
		mw := Middleware{
			Name: strings.TrimSuffix(filepath.Base(file), ".php"),
			Path: file,
		}
		middleware = append(middleware, mw)
	}

	return middleware, nil
}

func (a *Analyzer) analyzeProviders() ([]Provider, error) {
	providersPath := filepath.Join(a.projectPath, "app", "Providers")
	files, err := filepath.Glob(filepath.Join(providersPath, "*.php"))
	if err != nil {
		return nil, err
	}

	var providers []Provider
	for _, file := range files {
		provider := Provider{
			Name: strings.TrimSuffix(filepath.Base(file), ".php"),
			Path: file,
		}
		providers = append(providers, provider)
	}

	return providers, nil
}