package docker

import (
	"fmt"
	"path/filepath"
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
	sb.WriteString("skinparam component {\n")
	sb.WriteString("  BackgroundColor White\n")
	sb.WriteString("  ArrowColor Gray\n")
	sb.WriteString("  BorderColor Gray\n")
	sb.WriteString("}\n\n")

	// Título
	sb.WriteString(fmt.Sprintf("title Docker Project: %s\n\n", g.project.Name))

	// Gera diagrama para Dockerfiles
	if len(g.project.Dockerfiles) > 0 {
		sb.WriteString("package \"Dockerfiles\" {\n")
		for _, dockerfile := range g.project.Dockerfiles {
			g.generateDockerfile(&sb, dockerfile)
		}
		sb.WriteString("}\n\n")
	}

	// Gera diagrama para docker-compose
	if g.project.Compose != nil {
		g.generateCompose(&sb)
	}

	sb.WriteString("@enduml\n")
	return sb.String()
}

func (g *PlantUMLGenerator) generateDockerfile(sb *strings.Builder, dockerfile Dockerfile) {
	name := filepath.Base(dockerfile.Path)
	sb.WriteString(fmt.Sprintf("component \"%s\" as %s {\n", name, sanitizeName(name)))

	// Base image
	sb.WriteString(fmt.Sprintf("  note \"FROM %s\" as N1\n", dockerfile.Base))

	// Estágios
	if len(dockerfile.Stages) > 0 {
		for i, stage := range dockerfile.Stages {
			stageName := sanitizeName(fmt.Sprintf("%s_stage_%d", name, i))
			sb.WriteString(fmt.Sprintf("  component \"%s\" as %s {\n", stage.Name, stageName))
			sb.WriteString(fmt.Sprintf("    note \"FROM %s\" as N%d\n", stage.Base, i+2))

			// Steps do estágio
			if len(stage.Steps) > 0 {
				sb.WriteString("    note \"Steps\\n")
				for _, step := range stage.Steps {
					sb.WriteString(fmt.Sprintf("%s %s\\n", step.Type, escapeString(step.Command)))
				}
				sb.WriteString("\" as NS" + fmt.Sprintf("%d\n", i+1))
			}
			sb.WriteString("  }\n")
		}
	}

	// Environment variables
	if len(dockerfile.Env) > 0 {
		sb.WriteString("  note \"Environment\\n")
		for _, env := range dockerfile.Env {
			sb.WriteString(fmt.Sprintf("%s=%s\\n", env.Key, escapeString(env.Value)))
		}
		sb.WriteString("\" as NENV\n")
	}

	// Exposed ports
	if len(dockerfile.Expose) > 0 {
		sb.WriteString("  note \"Exposed Ports\\n")
		for _, port := range dockerfile.Expose {
			sb.WriteString(fmt.Sprintf("%s\\n", port))
		}
		sb.WriteString("\" as NPORTS\n")
	}

	// Volumes
	if len(dockerfile.Volumes) > 0 {
		sb.WriteString("  note \"Volumes\\n")
		for _, volume := range dockerfile.Volumes {
			sb.WriteString(fmt.Sprintf("%s\\n", volume))
		}
		sb.WriteString("\" as NVOLS\n")
	}

	// Commands (CMD/ENTRYPOINT)
	if len(dockerfile.Commands) > 0 {
		sb.WriteString("  note \"Commands\\n")
		for _, cmd := range dockerfile.Commands {
			sb.WriteString(fmt.Sprintf("%s %s\\n", cmd.Type, escapeString(cmd.Command)))
		}
		sb.WriteString("\" as NCMD\n")
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateCompose(sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("package \"Docker Compose (v%s)\" {\n", g.project.Compose.Version))

	// Gera cada serviço
	for name, service := range g.project.Compose.Services {
    service.Name = name // atribui a chave do map ao campo Name
    g.generateService(sb, service)
}

	// Gera networks
	if len(g.project.Compose.Networks) > 0 {
		sb.WriteString("\npackage \"Networks\" {\n")
		for _, network := range g.project.Compose.Networks {
			g.generateNetwork(sb, network)
		}
		sb.WriteString("}\n")
	}

	// Gera volumes
	if len(g.project.Compose.Volumes) > 0 {
		sb.WriteString("\npackage \"Volumes\" {\n")
		for _, volume := range g.project.Compose.Volumes {
			g.generateVolume(sb, volume)
		}
		sb.WriteString("}\n")
	}

	// Gera relacionamentos entre serviços
	for _, service := range g.project.Compose.Services {
		// DependsOn
		for _, dep := range service.DependsOn {
			sb.WriteString(fmt.Sprintf("%s ..> %s : depends on\n",
				sanitizeName(service.Name),
				sanitizeName(dep)))
		}

		// Networks
		for _, net := range service.Networks {
			sb.WriteString(fmt.Sprintf("%s -- %s\n",
				sanitizeName(service.Name),
				sanitizeName(net)))
		}

		// Volumes
		for _, vol := range service.Volumes {
			// Converte o volume para VolumeMapping
			volMapping := parseVolume(vol)
			// Se for bind mount, ignora
			if strings.HasPrefix(volMapping.Source, "./") || strings.HasPrefix(volMapping.Source, "/") {
				continue
			}
			sb.WriteString(fmt.Sprintf("%s -- %s\n",
				sanitizeName(service.Name),
				sanitizeName(volMapping.Source)))
		}
	}

	sb.WriteString("}\n\n")
}

func (g *PlantUMLGenerator) generateService(sb *strings.Builder, service Service) {
	serviceName := sanitizeName(service.Name)
	sb.WriteString(fmt.Sprintf("component \"%s\" as %s {\n", service.Name, serviceName))

	// Image ou build
	if service.Image != "" {
		sb.WriteString(fmt.Sprintf("  note \"Image: %s\" as N_%s_1\n", service.Image, serviceName))
	} else if service.Build != nil {
		sb.WriteString(fmt.Sprintf("  note \"Build:\\nContext: %s\\nDockerfile: %s\" as N_%s_1\n",
			service.Build.Context,
			service.Build.Dockerfile,
			serviceName))
	}

	// Ports
	if len(service.Ports) > 0 {
		sb.WriteString("  note \"Ports:\\n")
		for _, p := range service.Ports {
			portMapping := parsePort(p)
			if portMapping.Host != "" {
				sb.WriteString(fmt.Sprintf("%s:%s\\n", portMapping.Host, portMapping.Container))
			} else {
				sb.WriteString(fmt.Sprintf("%s\\n", portMapping.Container))
			}
		}
		sb.WriteString(fmt.Sprintf("\" as N_%s_2\n", serviceName))
	}

	// Environment
	if len(service.Environment) > 0 {
		sb.WriteString("  note \"Environment:\\n")
		for _, e := range service.Environment {
			envVar := parseEnv(e)
			sb.WriteString(fmt.Sprintf("%s=%s\\n", envVar.Key, escapeString(envVar.Value)))
		}
		sb.WriteString(fmt.Sprintf("\" as N_%s_3\n", serviceName))
	}

	sb.WriteString("}\n")
}

func (g *PlantUMLGenerator) generateNetwork(sb *strings.Builder, network Network) {
	networkName := sanitizeName(network.Name)
	sb.WriteString(fmt.Sprintf("component \"%s\" as %s <<network>> {\n", network.Name, networkName))
	if network.Driver != "" {
		sb.WriteString(fmt.Sprintf("  note \"Driver: %s\" as N_%s\n", network.Driver, networkName))
	}
	sb.WriteString("}\n")
}

func (g *PlantUMLGenerator) generateVolume(sb *strings.Builder, volume ComposeVolume) {
	volumeName := sanitizeName(volume.Name)
	sb.WriteString(fmt.Sprintf("database \"%s\" as %s {\n", volume.Name, volumeName))
	if volume.Driver != "" {
		sb.WriteString(fmt.Sprintf("  note \"Driver: %s\" as N_%s\n", volume.Driver, volumeName))
	}
	sb.WriteString("}\n")
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	return name
}

func escapeString(s string) string {
    // Substitui " por \\\" para que o PlantUML receba duas barras invertidas
    s = strings.ReplaceAll(s, "\"", "\\\\\"")
    s = strings.ReplaceAll(s, "\n", "\\n")
    return s
}
