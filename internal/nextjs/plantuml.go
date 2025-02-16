// internal/nextjs/plantuml.go
package nextjs

import (
	"fmt"
	"regexp"
	"strings"
)

type PlantUMLGenerator struct {
	project *Project
}

func NewPlantUMLGenerator(project *Project) *PlantUMLGenerator {
	return &PlantUMLGenerator{project: project}
}

func (g *PlantUMLGenerator) Generate() string {
	var sb strings.Builder

	// Cabeçalho do diagrama
	sb.WriteString("@startuml\n\n")
	sb.WriteString("!theme plain\n")
	sb.WriteString("skinparam linetype ortho\n")
	sb.WriteString("skinparam roundcorner 5\n")
	sb.WriteString("skinparam shadowing false\n")
	sb.WriteString("skinparam class {\n")
	sb.WriteString("  BackgroundColor White\n")
	sb.WriteString("  ArrowColor Gray\n")
	sb.WriteString("  BorderColor Gray\n")
	sb.WriteString("}\n\n")

	// Título
	sb.WriteString(fmt.Sprintf("title Next.js Project: %s\n\n", g.project.Name))

	// Componentes
	sb.WriteString("package Components {\n")
	for _, component := range g.project.Components {
		g.generateComponent(&sb, component)
	}
	sb.WriteString("}\n\n")

	// Páginas
	sb.WriteString("package Pages {\n")
	for _, page := range g.project.Pages {
		g.generatePage(&sb, page)
	}
	sb.WriteString("}\n\n")

	// Layouts
	if len(g.project.Layouts) > 0 {
		sb.WriteString("package Layouts {\n")
		for _, layout := range g.project.Layouts {
			g.generateLayout(&sb, layout)
		}
		sb.WriteString("}\n\n")
	}

	// Estado
	if len(g.project.StateModules) > 0 {
		sb.WriteString("package State {\n")
		for _, module := range g.project.StateModules {
			g.generateStateModule(&sb, module)
		}
		sb.WriteString("}\n\n")
	}

	// APIs
	if len(g.project.APIs) > 0 {
		sb.WriteString("package APIs {\n")
		for _, api := range g.project.APIs {
			g.generateAPI(&sb, api)
		}
		sb.WriteString("}\n\n")
	}

	// Gera relacionamentos
	g.generateRelationships(&sb)

	sb.WriteString("@enduml\n")
	return sb.String()
}

func (g *PlantUMLGenerator) generateComponent(sb *strings.Builder, component Component) {
	// Define ícone baseado no tipo de componente
	icon := "C"
	if component.IsServer {
		icon = "S"
	} else if component.IsClient {
		icon = "B" // Browser
	}

	sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (%s,#ADD1B2) component >> {\n",
		component.Name,
		sanitizeName(component.Name),
		icon))

	// Props
	if len(component.Props) > 0 {
		sb.WriteString("  .. props ..\n")
		for _, prop := range component.Props {
			required := ""
			if prop.Required {
				required = "*"
			}
			sb.WriteString(fmt.Sprintf("  %s%s: %s\n", required, prop.Name, prop.Type))
		}
	}

	// Hooks
	if len(component.Hooks) > 0 {
		sb.WriteString("  .. hooks ..\n")
		for _, hook := range component.Hooks {
			if len(hook.Dependencies) > 0 {
				sb.WriteString(fmt.Sprintf("  %s(%s)\n", hook.Name, strings.Join(hook.Dependencies, ", ")))
			} else {
				sb.WriteString(fmt.Sprintf("  %s()\n", hook.Name))
			}
		}
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generatePage(sb *strings.Builder, page Page) {
	icon := "P"
	if page.IsStatic {
		icon = "S"
	} else if page.IsDynamic {
		icon = "D"
	}

	sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (%s,#87CEFA) page >> {\n",
		page.Route,
		sanitizeName(page.Path),
		icon))

	// Parâmetros da rota
	if len(page.Params) > 0 {
		sb.WriteString("  .. params ..\n")
		for _, param := range page.Params {
			required := ""
			if param.Required {
				required = "*"
			}
			sb.WriteString(fmt.Sprintf("  %s%s: %s\n", required, param.Name, param.Type))
		}
	}

	// Componentes usados
	if len(page.Components) > 0 {
		sb.WriteString("  .. components ..\n")
		for _, comp := range page.Components {
			sb.WriteString(fmt.Sprintf("  + %s\n", comp))
		}
	}

	// APIs usadas
	if len(page.APIs) > 0 {
		sb.WriteString("  .. apis ..\n")
		for _, api := range page.APIs {
			sb.WriteString(fmt.Sprintf("  # %s\n", api))
		}
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateLayout(sb *strings.Builder, layout Layout) {
	icon := "L"
	if layout.IsRoot {
		icon = "R"
	}

	sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (%s,#B0C4DE) layout >> {\n",
		layout.Name,
		sanitizeName(layout.Name),
		icon))

	if len(layout.Components) > 0 {
		sb.WriteString("  .. components ..\n")
		for _, comp := range layout.Components {
			sb.WriteString(fmt.Sprintf("  + %s\n", comp))
		}
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateStateModule(sb *strings.Builder, module StateModule) {
	icon := "S"
	switch module.Type {
	case "redux":
		icon = "R"
	case "zustand":
		icon = "Z"
	case "jotai":
		icon = "J"
	}

	sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (%s,#FFB6C1) state >> {\n",
		module.Name,
		sanitizeName(module.Name),
		icon))

	// Actions
	if len(module.Actions) > 0 {
		sb.WriteString("  .. actions ..\n")
		for _, action := range module.Actions {
			if action.Payload != "" {
				sb.WriteString(fmt.Sprintf("  + %s(%s)\n", action.Name, action.Payload))
			} else {
				sb.WriteString(fmt.Sprintf("  + %s()\n", action.Name))
			}
		}
	}

	// Slices
	if len(module.Slices) > 0 {
		sb.WriteString("  .. slices ..\n")
		for _, slice := range module.Slices {
			sb.WriteString(fmt.Sprintf("  # %s: %s\n", slice.Name, slice.State))
			for _, reducer := range slice.Reducers {
				sb.WriteString(fmt.Sprintf("    - %s\n", reducer))
			}
		}
	}

	// Atoms
	if len(module.Atoms) > 0 {
		sb.WriteString("  .. atoms ..\n")
		for _, atom := range module.Atoms {
			if atom.Default != "" {
				sb.WriteString(fmt.Sprintf("  * %s: %s = %s\n", atom.Name, atom.Type, atom.Default))
			} else {
				sb.WriteString(fmt.Sprintf("  * %s: %s\n", atom.Name, atom.Type))
			}
		}
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateAPI(sb *strings.Builder, api API) {
	sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (A,#98FB98) api >> {\n",
		api.Route,
		sanitizeName(api.Route)))

	sb.WriteString(fmt.Sprintf("  %s %s\n", api.Method, api.Route))
	sb.WriteString(fmt.Sprintf("  handler: %s\n", api.Handler))

	if len(api.Middleware) > 0 {
		sb.WriteString("  .. middleware ..\n")
		for _, mw := range api.Middleware {
			sb.WriteString(fmt.Sprintf("  # %s\n", mw))
		}
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateRelationships(sb *strings.Builder) {
	// Página -> Layout
	for _, page := range g.project.Pages {
		if page.Layout != "" {
			sb.WriteString(fmt.Sprintf("%s ..> %s : uses layout\n",
				sanitizeName(page.Path),
				sanitizeName(page.Layout)))
		}
	}

	// Página -> Componentes
	for _, page := range g.project.Pages {
		for _, comp := range page.Components {
			sb.WriteString(fmt.Sprintf("%s ..> %s : uses\n",
				sanitizeName(page.Path),
				sanitizeName(comp)))
		}
	}

	// Página -> APIs
	for _, page := range g.project.Pages {
		for _, api := range page.APIs {
			sb.WriteString(fmt.Sprintf("%s ..> %s : calls\n",
				sanitizeName(page.Path),
				sanitizeName(api)))
		}
	}

	// Componente -> Componente (children)
	for _, comp := range g.project.Components {
		for _, child := range comp.Children {
			sb.WriteString(fmt.Sprintf("%s --* %s : contains\n",
				sanitizeName(comp.Name),
				sanitizeName(child.Name)))
		}
	}
}

func sanitizeName(name string) string {
	// Remove caracteres problemáticos para o PlantUML
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, " ", "_")
	
	// Remove parâmetros dinâmicos
	re := regexp.MustCompile(`\[([^\]]+)\]`)
	name = re.ReplaceAllString(name, "param")

	// Garante que começa com letra ou underscore
	if len(name) > 0 && !strings.Contains("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_", string(name[0])) {
		name = "_" + name
	}

	return name
}