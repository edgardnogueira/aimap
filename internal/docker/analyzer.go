// internal/docker/analyzer.go
package docker

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
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

	// Analisa Dockerfiles
	dockerfiles, err := a.findDockerfiles()
	if err != nil {
		slog.Error("Erro ao procurar Dockerfiles", "error", err)
	} else {
		for _, file := range dockerfiles {
			dockerfile, err := a.analyzeDockerfile(file)
			if err != nil {
				slog.Error("Erro ao analisar Dockerfile", 
					"file", file, 
					"error", err)
				continue
			}
			project.Dockerfiles = append(project.Dockerfiles, *dockerfile)
		}
	}

	// Analisa docker-compose
	compose, err := a.analyzeCompose()
	if err != nil {
		slog.Error("Erro ao analisar docker-compose", "error", err)
	} else {
		project.Compose = compose
	}

	return project, nil
}

func (a *Analyzer) findDockerfiles() ([]string, error) {
	var files []string

	// Procura por Dockerfile e Dockerfile.*
	matches, err := filepath.Glob(filepath.Join(a.projectPath, "Dockerfile*"))
	if err != nil {
		return nil, err
	}
	files = append(files, matches...)

	// Procura em subdiretórios
	err = filepath.Walk(a.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "Dockerfile") {
			// Ignora se já encontrado
			for _, f := range matches {
				if f == path {
					return nil
				}
			}
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func (a *Analyzer) analyzeDockerfile(path string) (*Dockerfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dockerfile := &Dockerfile{
		Path: path,
	}

	var currentStage *Stage
	scanner := bufio.NewScanner(file)
	
	// Regex para comandos do Dockerfile
	fromRe := regexp.MustCompile(`^FROM\s+(?:--platform=\S+\s+)?(\S+)(?:\s+(?:as|AS)\s+(\S+))?`)
	runRe := regexp.MustCompile(`^RUN\s+(.+)`)
	copyRe := regexp.MustCompile(`^COPY\s+(?:--from=\S+\s+)?(.+)`)
	addRe := regexp.MustCompile(`^ADD\s+(.+)`)
	envRe := regexp.MustCompile(`^ENV\s+(\S+)(?:\s+|\=)(.+)`)
	exposeRe := regexp.MustCompile(`^EXPOSE\s+(.+)`)
	volumeRe := regexp.MustCompile(`^VOLUME\s+(.+)`)
	cmdRe := regexp.MustCompile(`^CMD\s+(.+)`)
	entrypointRe := regexp.MustCompile(`^ENTRYPOINT\s+(.+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// FROM inicia um novo estágio
		if matches := fromRe.FindStringSubmatch(line); matches != nil {
			// Se é o primeiro FROM, define a imagem base
			if dockerfile.Base == "" {
				dockerfile.Base = matches[1]
			}

			// Se tem nome de estágio, cria um novo
			if len(matches) > 2 && matches[2] != "" {
				currentStage = &Stage{
					Name: matches[2],
					Base: matches[1],
				}
				dockerfile.Stages = append(dockerfile.Stages, *currentStage)
			} else {
				currentStage = nil
			}
			continue
		}

		// Processa outros comandos
		switch {
		case runRe.MatchString(line):
			matches := runRe.FindStringSubmatch(line)
			step := Step{
				Type:    "RUN",
				Command: matches[1],
			}
			if currentStage != nil {
				currentStage.Steps = append(currentStage.Steps, step)
			} else {
				dockerfile.Steps = append(dockerfile.Steps, step)
			}

		case copyRe.MatchString(line):
			matches := copyRe.FindStringSubmatch(line)
			step := Step{
				Type:    "COPY",
				Command: matches[1],
			}
			if currentStage != nil {
				currentStage.Steps = append(currentStage.Steps, step)
			} else {
				dockerfile.Steps = append(dockerfile.Steps, step)
			}

		case addRe.MatchString(line):
			matches := addRe.FindStringSubmatch(line)
			step := Step{
				Type:    "ADD",
				Command: matches[1],
			}
			if currentStage != nil {
				currentStage.Steps = append(currentStage.Steps, step)
			} else {
				dockerfile.Steps = append(dockerfile.Steps, step)
			}

		case envRe.MatchString(line):
			matches := envRe.FindStringSubmatch(line)
			env := EnvVar{
				Key:   strings.TrimSpace(matches[1]),
				Value: strings.TrimSpace(matches[2]),
			}
			dockerfile.Env = append(dockerfile.Env, env)

		case exposeRe.MatchString(line):
			matches := exposeRe.FindStringSubmatch(line)
			ports := strings.Fields(matches[1])
			dockerfile.Expose = append(dockerfile.Expose, ports...)

		case volumeRe.MatchString(line):
			matches := volumeRe.FindStringSubmatch(line)
			volumes := strings.Fields(matches[1])
			dockerfile.Volumes = append(dockerfile.Volumes, volumes...)

		case cmdRe.MatchString(line):
			matches := cmdRe.FindStringSubmatch(line)
			dockerfile.Commands = append(dockerfile.Commands, Command{
				Type:    "CMD",
				Command: matches[1],
			})

		case entrypointRe.MatchString(line):
			matches := entrypointRe.FindStringSubmatch(line)
			dockerfile.Commands = append(dockerfile.Commands, Command{
				Type:    "ENTRYPOINT",
				Command: matches[1],
			})
		}
	}

	return dockerfile, scanner.Err()
}

func (a *Analyzer) analyzeCompose() (*Compose, error) {
	// Procura por docker-compose.yml ou docker-compose.yaml
	composeFiles := []string{
		filepath.Join(a.projectPath, "docker-compose.yml"),
		filepath.Join(a.projectPath, "docker-compose.yaml"),
	}

	var composeFile string
	for _, file := range composeFiles {
		if _, err := os.Stat(file); err == nil {
			composeFile = file
			break
		}
	}

	if composeFile == "" {
		return nil, fmt.Errorf("arquivo docker-compose não encontrado")
	}

	data, err := os.ReadFile(composeFile)
	if err != nil {
		return nil, err
	}

	var compose Compose
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, err
	}

	return &compose, nil
}