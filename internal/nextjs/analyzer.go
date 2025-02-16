// internal/nextjs/analyzer.go
package nextjs

import (
	"bytes"
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

	// Analisa componentes
	components, err := a.analyzeComponents()
	if err != nil {
		slog.Error("Erro ao analisar componentes", "error", err)
	}
	project.Components = components

	// Analisa páginas
	pages, err := a.analyzePages()
	if err != nil {
		slog.Error("Erro ao analisar páginas", "error", err)
	}
	project.Pages = pages

	// Analisa layouts
	layouts, err := a.analyzeLayouts()
	if err != nil {
		slog.Error("Erro ao analisar layouts", "error", err)
	}
	project.Layouts = layouts

	// Analisa módulos de estado
	stateModules, err := a.analyzeStateModules()
	if err != nil {
		slog.Error("Erro ao analisar módulos de estado", "error", err)
	}
	project.StateModules = stateModules

	// Analisa APIs
	apis, err := a.analyzeAPIs()
	if err != nil {
		slog.Error("Erro ao analisar APIs", "error", err)
	}
	project.APIs = apis

	return project, nil
}

func (a *Analyzer) analyzeComponents() ([]Component, error) {
	var components []Component

	// Procura em src/components e components
	componentsPaths := []string{
		filepath.Join(a.projectPath, "src", "components"),
		filepath.Join(a.projectPath, "components"),
	}

	for _, path := range componentsPaths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && isComponentFile(info.Name()) {
				component, err := a.analyzeComponentFile(path)
				if err != nil {
					slog.Error("Erro ao analisar componente",
						"file", path,
						"error", err)
					return nil
				}
				components = append(components, *component)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	return components, nil
}

func (a *Analyzer) analyzeComponentFile(path string) (*Component, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	component := &Component{
		Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Path: path,
		Type: "functional", // Assume functional por padrão
	}

	// Verifica se é um componente do lado do cliente
	clientRe := regexp.MustCompile(`"use client"`)
	component.IsClient = clientRe.Match(content)

	// Verifica se é um componente do lado do servidor
	serverRe := regexp.MustCompile(`"use server"`)
	component.IsServer = serverRe.Match(content)

	// Procura props usando TypeScript interface/type
	propsRe := regexp.MustCompile(`(?:interface|type)\s+(\w+Props)\s*\{([^}]+)\}`)
	if matches := propsRe.FindSubmatch(content); len(matches) > 2 {
		props := parseProps(string(matches[2]))
		component.Props = props
	}

	// Procura hooks
	hooksRe := regexp.MustCompile(`use\w+\(([^)]*)\)`)
	matches := hooksRe.FindAllSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			hookName := string(match[0])
			hookType := strings.TrimSuffix(strings.TrimPrefix(hookName, "use"), "(")
			deps := []string{}
			if len(match) > 1 {
				depsStr := string(match[1])
				deps = strings.Split(depsStr, ",")
			}
			
			component.Hooks = append(component.Hooks, Hook{
				Name: hookName,
				Type: hookType,
				Dependencies: deps,
			})
		}
	}

	return component, nil
}

func (a *Analyzer) analyzePages() ([]Page, error) {
	var pages []Page

	// Procura em src/app e app (Next.js 13+) e pages (Next.js < 13)
	pagesPaths := []string{
		filepath.Join(a.projectPath, "src", "app"),
		filepath.Join(a.projectPath, "app"),
		filepath.Join(a.projectPath, "src", "pages"),
		filepath.Join(a.projectPath, "pages"),
	}

	for _, path := range pagesPaths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && isPageFile(info.Name()) {
				page, err := a.analyzePageFile(path)
				if err != nil {
					slog.Error("Erro ao analisar página",
						"file", path,
						"error", err)
					return nil
				}
				pages = append(pages, *page)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	return pages, nil
}

func (a *Analyzer) analyzeStateModules() ([]StateModule, error) {
	var modules []StateModule

	// Procura em src/store, store, src/state, state
	statePaths := []string{
		filepath.Join(a.projectPath, "src", "store"),
		filepath.Join(a.projectPath, "store"),
		filepath.Join(a.projectPath, "src", "state"),
		filepath.Join(a.projectPath, "state"),
	}

	for _, path := range statePaths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && isStateFile(info.Name()) {
				module, err := a.analyzeStateFile(path)
				if err != nil {
					slog.Error("Erro ao analisar módulo de estado",
						"file", path,
						"error", err)
					return nil
				}
				modules = append(modules, *module)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	return modules, nil
}

func isComponentFile(name string) bool {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	
	validExts := map[string]bool{".tsx": true, ".jsx": true, ".js": true}
	if !validExts[ext] {
		return false
	}

	// Componentes geralmente começam com letra maiúscula
	if len(base) > 0 && !strings.Contains("ABCDEFGHIJKLMNOPQRSTUVWXYZ", string(base[0])) {
		return false
	}

	return true
}

func isPageFile(name string) bool {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	
	validExts := map[string]bool{".tsx": true, ".jsx": true, ".js": true}
	if !validExts[ext] {
		return false
	}

	validNames := map[string]bool{
		"page": true,
		"index": true,
		"_app": true,
		"_document": true,
	}

	return validNames[base]
}

func isStateFile(name string) bool {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	
	validExts := map[string]bool{".ts": true, ".js": true}
	if !validExts[ext] {
		return false
	}

	// Arquivos de estado comuns
	stateIndicators := []string{
		"store", "slice", "reducer", "atom", "state",
	}

	for _, indicator := range stateIndicators {
		if strings.Contains(strings.ToLower(base), indicator) {
			return true
		}
	}

	return false
}

func parseProps(propsStr string) []Prop {
	var props []Prop
	lines := strings.Split(propsStr, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Procura por "name: type" ou "name?: type"
		re := regexp.MustCompile(`(\w+)(\??)\s*:\s*([^;]+)`)
		if matches := re.FindStringSubmatch(line); len(matches) > 3 {
			props = append(props, Prop{
				Name:     matches[1],
				Type:     matches[3],
				Required: matches[2] != "?",
			})
		}
	}

	return props
}


// internal/nextjs/analyzer.go (continuação)

func (a *Analyzer) analyzeLayouts() ([]Layout, error) {
    var layouts []Layout

    // Procura em src/app/layout.tsx e app/layout.tsx (Next.js 13+)
    layoutPaths := []string{
        filepath.Join(a.projectPath, "src", "app", "layout.tsx"),
        filepath.Join(a.projectPath, "app", "layout.tsx"),
    }

    for _, path := range layoutPaths {
        if _, err := os.Stat(path); err == nil {
            content, err := ioutil.ReadFile(path)
            if err != nil {
                slog.Error("Erro ao ler arquivo de layout", "file", path, "error", err)
                continue
            }

            layout := Layout{
                Name: "RootLayout",
                Path: path,
                IsRoot: true,
            }

            // Busca por componentes importados
            importRe := regexp.MustCompile(`import\s+(\w+)\s+from\s+['"]([^'"]+)['"]`)
            matches := importRe.FindAllSubmatch(content, -1)
            for _, match := range matches {
                if len(match) > 1 {
                    layout.Components = append(layout.Components, string(match[1]))
                }
            }

            layouts = append(layouts, layout)
        }
    }

    return layouts, nil
}

func (a *Analyzer) analyzePageFile(path string) (*Page, error) {
    content, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    relPath, err := filepath.Rel(a.projectPath, path)
    if err != nil {
        relPath = path
    }

    page := &Page{
        Path: relPath,
        Route: pathToRoute(relPath),
        IsStatic: true, // Assume estático por padrão
    }

    // Verifica se é uma página dinâmica
    if strings.Contains(relPath, "[") && strings.Contains(relPath, "]") {
        page.IsStatic = false
        page.IsDynamic = true

        // Extrai parâmetros dinâmicos
        paramRe := regexp.MustCompile(`\[([^\]]+)\]`)
        matches := paramRe.FindAllStringSubmatch(relPath, -1)
        for _, match := range matches {
            if len(match) > 1 {
                param := Param{
                    Name: match[1],
                    Type: "string",
                    Required: !strings.HasPrefix(match[1], "...") && !strings.HasSuffix(match[1], "?"),
                }
                page.Params = append(page.Params, param)
            }
        }
    }

    // Busca por componentes importados
    importRe := regexp.MustCompile(`import\s+(\w+)\s+from\s+['"]([^'"]+)['"]`)
    matches := importRe.FindAllSubmatch(content, -1)
    for _, match := range matches {
        if len(match) > 1 {
            page.Components = append(page.Components, string(match[1]))
        }
    }

    // Busca por chamadas de API
    apiRe := regexp.MustCompile(`(fetch|axios\.get|axios\.post|axios\.put|axios\.delete)\s*\(\s*['"]([^'"]+)['"]`)
    matches = apiRe.FindAllSubmatch(content, -1)
    for _, match := range matches {
        if len(match) > 2 {
            page.APIs = append(page.APIs, string(match[2]))
        }
    }

    return page, nil
}

func (a *Analyzer) analyzeAPIs() ([]API, error) {
    var apis []API

    // Procura em src/app/api e app/api (Next.js 13+) ou src/pages/api e pages/api
    apiPaths := []string{
        filepath.Join(a.projectPath, "src", "app", "api"),
        filepath.Join(a.projectPath, "app", "api"),
        filepath.Join(a.projectPath, "src", "pages", "api"),
        filepath.Join(a.projectPath, "pages", "api"),
    }

    for _, basePath := range apiPaths {
        err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }

            if !info.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".js")) {
                content, err := ioutil.ReadFile(path)
                if err != nil {
                    slog.Error("Erro ao ler arquivo API", "file", path, "error", err)
                    return nil
                }

                relPath, err := filepath.Rel(basePath, path)
                if err != nil {
                    relPath = path
                }

                api := API{
                    Route: "/api/" + strings.TrimSuffix(relPath, filepath.Ext(relPath)),
                    Path: path,
                    Method: detectHTTPMethod(content),
                    Handler: detectHandler(content),
                }

                // Busca por middleware
                middlewareRe := regexp.MustCompile(`use\(([^)]+)\)`)
                matches := middlewareRe.FindAllSubmatch(content, -1)
                for _, match := range matches {
                    if len(match) > 1 {
                        api.Middleware = append(api.Middleware, string(match[1]))
                    }
                }

                apis = append(apis, api)
            }
            return nil
        })
        if err != nil && !os.IsNotExist(err) {
            return nil, err
        }
    }

    return apis, nil
}

func (a *Analyzer) analyzeStateFile(path string) (*StateModule, error) {
    content, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    module := &StateModule{
        Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
        Path: path,
    }

    // Determina o tipo de estado
    if bytes.Contains(content, []byte("createSlice")) {
        module.Type = "redux"
    } else if bytes.Contains(content, []byte("create((set)")) {
        module.Type = "zustand"
    } else if bytes.Contains(content, []byte("atom(")) {
        module.Type = "jotai"
    } else {
        module.Type = "custom"
    }

    switch module.Type {
    case "redux":
        // Procura por actions
        actionRe := regexp.MustCompile(`(\w+):\s*\((state,\s*action\)`)
        matches := actionRe.FindAllSubmatch(content, -1)
        for _, match := range matches {
            if len(match) > 1 {
                module.Actions = append(module.Actions, Action{
                    Name: string(match[1]),
                    Type: "reducer",
                })
            }
        }

        // Procura por slices
        sliceRe := regexp.MustCompile(`createSlice\s*\(\s*\{\s*name:\s*['"](\w+)['"]`)
        matches = sliceRe.FindAllSubmatch(content, -1)
        for _, match := range matches {
            if len(match) > 1 {
                module.Slices = append(module.Slices, Slice{
                    Name: string(match[1]),
                })
            }
        }

    case "zustand":
        // Procura por setters
        setterRe := regexp.MustCompile(`set\s*\(\s*\{\s*(\w+):\s*([^}]+)\s*\}`)
        matches := setterRe.FindAllSubmatch(content, -1)
        for _, match := range matches {
            if len(match) > 1 {
                module.Actions = append(module.Actions, Action{
                    Name: "set" + string(match[1]),
                    Type: "setter",
                })
            }
        }

    case "jotai":
        // Procura por atoms
        atomRe := regexp.MustCompile(`atom\s*\(\s*(\w+)\s*\)`)
        matches := atomRe.FindAllSubmatch(content, -1)
        for _, match := range matches {
            if len(match) > 1 {
                module.Atoms = append(module.Atoms, Atom{
                    Name: string(match[1]),
                    Type: "primitive",
                })
            }
        }
    }

    return module, nil
}

func pathToRoute(path string) string {
    // Remove extensão e index
    route := strings.TrimSuffix(path, filepath.Ext(path))
    route = strings.TrimSuffix(route, "/index")

    // Converte para formato de rota
    route = strings.ReplaceAll(route, "\\", "/")
    if !strings.HasPrefix(route, "/") {
        route = "/" + route
    }

    return route
}

func detectHTTPMethod(content []byte) string {
    // Detecta o método HTTP baseado no conteúdo
    methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
    for _, method := range methods {
        if bytes.Contains(content, []byte(method)) {
            return method
        }
    }
    return "GET" // Padrão
}

func detectHandler(content []byte) string {
    // Procura pelo nome da função handler
    handlerRe := regexp.MustCompile(`export\s+(?:default\s+)?(?:async\s+)?function\s+(\w+)`)
    if matches := handlerRe.FindSubmatch(content); len(matches) > 1 {
        return string(matches[1])
    }
    return "handler"
}