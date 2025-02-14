package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load carrega a configuração do arquivo especificado
func Load(configPath string) (*Config, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    if err := validateConfig(&cfg); err != nil {
        return nil, err
    }

    // Normaliza os paths
    cfg = normalizeConfig(cfg)

    return &cfg, nil
}

// validateConfig valida as configurações carregadas
func validateConfig(cfg *Config) error {
    if cfg.Output.Format == "" {
        return errors.New("formato de saída não especificado")
    }

    switch cfg.Output.Format {
    case "html", "markdown", "json", "yaml", "text":
        // formatos válidos
    default:
        return errors.New("formato de saída inválido (use: html, markdown, json, yaml ou text)")
    }

    if cfg.Output.Path == "" {
        cfg.Output.Path = "./docs" // valor padrão
    }

    // Valida se pelo menos um módulo está habilitado
    if !cfg.Golang.Enabled && !cfg.Kubernetes.Enabled && !cfg.Databases.Enabled {
        return errors.New("pelo menos um módulo (golang, kubernetes ou databases) deve estar habilitado")
    }

    // Valida configurações específicas quando habilitadas
    if cfg.Golang.Enabled {
        if len(cfg.Golang.Paths) == 0 {
            return errors.New("nenhum caminho especificado para análise de código Go")
        }
        if cfg.Golang.ReportLevel == "" {
            cfg.Golang.ReportLevel = "standard" // valor padrão
        }
        if err := validateReportLevel(cfg.Golang.ReportLevel); err != nil {
            return err
        }
    }

    if cfg.Kubernetes.Enabled {
        if len(cfg.Kubernetes.Paths) == 0 {
            return errors.New("nenhum caminho especificado para análise de recursos Kubernetes")
        }
    }

    if cfg.Databases.Enabled {
        if len(cfg.Databases.Connections) == 0 {
            return errors.New("nenhuma conexão de banco de dados configurada")
        }
        for _, conn := range cfg.Databases.Connections {
            if err := validateDatabaseConfig(conn); err != nil {
                return err
            }
        }
    }

    return nil
}

func validateReportLevel(level string) error {
    switch level {
    case "short", "standard", "complete":
        return nil
    default:
        return errors.New("nível de relatório inválido (use: short, standard ou complete)")
    }
}
// validateDatabaseConfig valida uma configuração específica de banco de dados
func validateDatabaseConfig(cfg DatabaseConfig) error {
    if cfg.Name == "" {
        return errors.New("nome da conexão de banco de dados não especificado")
    }
    if cfg.Type != "postgres" && cfg.Type != "mysql" {
        return errors.New("tipo de banco de dados inválido (use: postgres ou mysql)")
    }
    if cfg.Host == "" {
        return errors.New("host do banco de dados não especificado")
    }
    if cfg.Port == 0 {
        return errors.New("porta do banco de dados não especificada")
    }
    if cfg.Database == "" {
        return errors.New("nome do banco de dados não especificado")
    }
    return nil
}

// normalizeConfig normaliza os caminhos na configuração
func normalizeConfig(cfg Config) Config {
    // Normaliza o caminho de saída
    cfg.Output.Path = filepath.Clean(cfg.Output.Path)

    // Normaliza os caminhos Go
    for i, path := range cfg.Golang.Paths {
        cfg.Golang.Paths[i] = filepath.Clean(path)
    }

    // Normaliza os caminhos Kubernetes
    for i, path := range cfg.Kubernetes.Paths {
        cfg.Kubernetes.Paths[i] = filepath.Clean(path)
    }

    return cfg
}