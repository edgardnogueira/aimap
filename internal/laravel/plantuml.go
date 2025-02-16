// internal/laravel/plantuml.go
package laravel

import (
	"fmt"
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
	sb.WriteString(fmt.Sprintf("title Laravel Project: %s\n\n", g.project.Name))

	// Models
	sb.WriteString("package Models {\n")
	for _, model := range g.project.Models {
		g.generateModel(&sb, model)
	}
	sb.WriteString("}\n\n")

	// Controllers
	sb.WriteString("package Controllers {\n")
	for _, controller := range g.project.Controllers {
		g.generateController(&sb, controller)
	}
	sb.WriteString("}\n\n")

	// Middleware
	if len(g.project.Middleware) > 0 {
		sb.WriteString("package Middleware {\n")
		for _, mw := range g.project.Middleware {
			g.generateMiddleware(&sb, mw)
		}
		sb.WriteString("}\n\n")
	}

	// Service Providers
	if len(g.project.Providers) > 0 {
		sb.WriteString("package Providers {\n")
		for _, provider := range g.project.Providers {
			g.generateProvider(&sb, provider)
		}
		sb.WriteString("}\n\n")
	}

	// Relacionamentos
	g.generateRelationships(&sb)

	// Rotas
	g.generateRoutes(&sb)

	sb.WriteString("@enduml\n")
	return sb.String()
}

func (g *PlantUMLGenerator) generateModel(sb *strings.Builder, model Model) {
	sb.WriteString(fmt.Sprintf("class \"%s\" as %s {\n", model.Name, sanitizeName(model.Name)))
	
	// Tabela
	if model.Table != "" {
		sb.WriteString(fmt.Sprintf("  .. table: %s ..\n", model.Table))
	}

	// Fillable
	if len(model.Fillable) > 0 {
		sb.WriteString("  .. fillable ..\n")
	// internal/laravel/plantuml.go (continuação)

		for _, field := range model.Fillable {
			sb.WriteString(fmt.Sprintf("  + %s\n", field))
		}
	}

	// Hidden
	if len(model.Hidden) > 0 {
		sb.WriteString("  .. hidden ..\n")
		for _, field := range model.Hidden {
			sb.WriteString(fmt.Sprintf("  - %s\n", field))
		}
	}

	// Casts
	if len(model.Casts) > 0 {
		sb.WriteString("  .. casts ..\n")
		for _, cast := range model.Casts {
			sb.WriteString(fmt.Sprintf("  # %s : %s\n", cast.Field, cast.CastType))
		}
	}

	// Relationships
	if len(model.Relationships) > 0 {
		sb.WriteString("  .. relationships ..\n")
		for _, rel := range model.Relationships {
			sb.WriteString(fmt.Sprintf("  * %s() : %s\n", rel.Method, rel.Type))
		}
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateController(sb *strings.Builder, controller Controller) {
	sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (C,#ADD1B2) controller >> {\n",
		controller.Name, sanitizeName(controller.Name)))

	// Métodos
	for _, method := range controller.Methods {
		var params []string
		for _, param := range method.Parameters {
			params = append(params, sanitizeName(param))
		}
		sb.WriteString(fmt.Sprintf("  + %s(%s) : %s\n",
			method.Name,
			strings.Join(params, ", "),
			method.ReturnType))
	}

	// Middleware
	if len(controller.Middleware) > 0 {
		sb.WriteString("  .. middleware ..\n")
		for _, mw := range controller.Middleware {
			sb.WriteString(fmt.Sprintf("  # %s\n", mw))
		}
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateMiddleware(sb *strings.Builder, mw Middleware) {
	icon := "M"
	if mw.Global {
		icon = "G"
	}
	sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (%s,#B4A7E5) middleware >> {\n",
		mw.Name, sanitizeName(mw.Name), icon))
	
	if mw.Group != "" {
		sb.WriteString(fmt.Sprintf("  group: %s\n", mw.Group))
	}
	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateProvider(sb *strings.Builder, provider Provider) {
	sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (P,#85BBF0) provider >> {\n",
		provider.Name, sanitizeName(provider.Name)))
	
	sb.WriteString(fmt.Sprintf("  type: %s\n", provider.Type))
	if provider.Deferred {
		sb.WriteString("  deferred: true\n")
	}
	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateRelationships(sb *strings.Builder) {
	// Model relationships
	for _, model := range g.project.Models {
		modelName := sanitizeName(model.Name)
		for _, rel := range model.Relationships {
			relatedModel := sanitizeName(rel.RelatedModel)
			switch rel.Type {
			case "hasOne", "hasMany":
				sb.WriteString(fmt.Sprintf("%s \"1\" --> \"%s\" %s: %s\n",
					modelName,
					getMultiplicity(rel.Type),
					relatedModel,
					rel.Method))
			case "belongsTo":
				sb.WriteString(fmt.Sprintf("%s \"%s\" <-- \"1\" %s: %s\n",
					modelName,
					getMultiplicity(rel.Type),
					relatedModel,
					rel.Method))
			case "belongsToMany":
				sb.WriteString(fmt.Sprintf("%s \"*\" <--> \"*\" %s: %s\n",
					modelName,
					relatedModel,
					rel.Method))
				if rel.PivotTable != "" {
					sb.WriteString(fmt.Sprintf("note on link\n  pivot: %s\nend note\n",
						rel.PivotTable))
				}
			}
		}
	}

	// Controller para Model relationships
	for _, controller := range g.project.Controllers {
		controllerName := sanitizeName(controller.Name)
		// Assumindo que o nome do controller termina com "Controller"
		relatedModel := strings.TrimSuffix(controller.Name, "Controller")
		if modelExists(g.project.Models, relatedModel) {
			sb.WriteString(fmt.Sprintf("%s ..> %s: controls\n",
				controllerName,
				sanitizeName(relatedModel)))
		}
	}
}

func (g *PlantUMLGenerator) generateRoutes(sb *strings.Builder) {
    if len(g.project.Routes) > 0 {
        sb.WriteString("\npackage Routes {\n")
        for _, route := range g.project.Routes {
            routeName := route.Name
            if routeName == "" {
                routeName = route.URI
            }
            
            // Cria um identificador seguro para a rota
            routeId := "route_" + strings.NewReplacer(
                "/", "_",
                "{", "",
                "}", "",
                "-", "_",
                ".", "_",
                ":", "_",
                " ", "_",
            ).Replace(route.URI)

            // Escapa o URI para exibição
            displayURI := strings.ReplaceAll(route.URI, "{", "\\{")
            displayURI = strings.ReplaceAll(displayURI, "}", "\\}")
            
            sb.WriteString(fmt.Sprintf("class \"%s\" as %s << (R,#FFA07A) route >> {\n",
                displayURI,
                routeId))
            sb.WriteString(fmt.Sprintf("  %s %s\n", route.Method, displayURI))
            sb.WriteString(fmt.Sprintf("  action: %s\n", route.Action))
            if len(route.Middleware) > 0 {
                sb.WriteString("  .. middleware ..\n")
                for _, mw := range route.Middleware {
                    sb.WriteString(fmt.Sprintf("  # %s\n", mw))
                }
            }
            sb.WriteString("}\n")

            // Link route to controller
            if route.Action != "" {
                parts := strings.Split(route.Action, "@")
                if len(parts) == 2 {
                    controllerName := sanitizeName(parts[0])
                    sb.WriteString(fmt.Sprintf("%s ..> %s: %s\n",
                        routeId,
                        controllerName,
                        parts[1]))
                }
            }
        }
        sb.WriteString("}\n")
    }
}

func sanitizeName(name string) string {
	// Remove caracteres problemáticos para o PlantUML
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func getMultiplicity(relType string) string {
	switch relType {
	case "hasOne", "belongsTo":
		return "1"
	case "hasMany", "belongsToMany":
		return "*"
	default:
		return "1"
	}
}

func modelExists(models []Model, name string) bool {
	for _, model := range models {
		if model.Name == name {
			return true
		}
	}
	return false
}