package config

type Config struct {
    Output      OutputConfig      `yaml:"output"`
    Golang      GolangConfig      `yaml:"golang"`
    Kubernetes  KubernetesConfig  `yaml:"kubernetes"`
    Databases   DatabasesConfig   `yaml:"databases"`
}

type OutputConfig struct {
    Format string `yaml:"format"` // html, markdown, json, yaml
    Path   string `yaml:"path"`
}


type GolangConfig struct {
    Enabled       bool            `yaml:"enabled"`
    ReportLevel   string         `yaml:"report_level"`   // short, standard, complete
    ReportOptions ReportOptions  `yaml:"report_options"`
    Paths         []string       `yaml:"paths"`
    Ignores       []string       `yaml:"ignores"`
}

type ReportOptions struct {
    ShowImports       bool `yaml:"show_imports"`
    ShowInternalFuncs bool `yaml:"show_internal_funcs"`
    ShowTests         bool `yaml:"show_tests"`
    ShowExamples      bool `yaml:"show_examples"`
}

type KubernetesConfig struct {
    Enabled bool     `yaml:"enabled"`
    Paths   []string `yaml:"paths"`
    Ignores []string `yaml:"ignores"`
}

type DatabasesConfig struct {
    Enabled     bool              `yaml:"enabled"`
    Connections []DatabaseConfig  `yaml:"connections"`
}

type DatabaseConfig struct {
    Name     string `yaml:"name"`
    Type     string `yaml:"type"`     // postgres, mysql
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    Database string `yaml:"database"`
}