package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/edgardnogueira/aimap/internal/config"
	"github.com/edgardnogueira/aimap/internal/godoc"
	"github.com/edgardnogueira/aimap/internal/kubedoc"
	"github.com/edgardnogueira/aimap/internal/output"
)

func main() {
    // Configurar logging estruturado
    slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })))

    // Parse flags
    configFile := flag.String("config", "config.yml", "Caminho para o arquivo de configuração")
    flag.Parse()

    // Carregar configuração
    cfg, err := config.Load(*configFile)
    if err != nil {
        slog.Error("Erro ao carregar configuração", "error", err)
        os.Exit(1)
    }

    // Criar gerador de documentação
    generator := output.NewGenerator(cfg.Output.Format, cfg.Output.Path)

    // Documentação Go
     if cfg.Golang.Enabled {
        slog.Info("Gerando documentação Go")
        goAnalyzer := godoc.NewAnalyzer(cfg.Golang)
        docs, err := goAnalyzer.Analyze()
        if err != nil {
            slog.Error("Erro ao analisar código Go", "error", err)
        } else {
            // Passa a configuração Go para o gerador
            if err := generator.AddGoDocumentation(docs, cfg.Golang); err != nil {
                slog.Error("Erro ao adicionar documentação Go", "error", err)
            }
        }
    }

    // Documentação Kubernetes
    if cfg.Kubernetes.Enabled {
        slog.Info("Gerando documentação Kubernetes")
        k8sAnalyzer := kubedoc.NewAnalyzer(cfg.Kubernetes)
        docs, err := k8sAnalyzer.Analyze()
        if err != nil {
            slog.Error("Erro ao analisar recursos Kubernetes", "error", err)
        } else {
            if err := generator.AddKubernetesDocumentation(docs); err != nil {
                slog.Error("Erro ao adicionar documentação Kubernetes", "error", err)
            }
        }
    }

    // Gerar documentação final
    if err := generator.Generate(); err != nil {
        slog.Error("Erro ao gerar documentação", "error", err)
        os.Exit(1)
    }

    slog.Info("Documentação gerada com sucesso", "output_path", cfg.Output.Path)
}