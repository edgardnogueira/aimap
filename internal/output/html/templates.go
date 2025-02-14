package html

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/edgardnogueira/aimap/internal/config"
)

// Template representa um template HTML
type Template struct {
    tmpl *template.Template
}

// ReportConfiguration representa a configuração do relatório
type ReportConfiguration struct {
    ReportLevel    string
    ReportOptions  config.ReportOptions
}

// TemplateData representa os dados passados para o template
type TemplateData struct {
    Go      interface{}
    K8s     interface{}
    Config  ReportConfiguration
    GoMermaid   string   
}

// NewTemplate cria um novo template HTML
func NewTemplate() *Template {
    t := template.New("documentation").Funcs(template.FuncMap{
        "safe": func(s string) template.HTML {
            return template.HTML(s)
        },
        "join": strings.Join,
        "shouldShow": func(data TemplateData, contentType string) bool {
            switch data.Config.ReportLevel {
            case "short":
                return contentType == "structs" || contentType == "interfaces"
            case "standard", "complete":
                return true
            default:
                return true
            }
        },
        "shouldShowImports": func(data TemplateData) bool {
            return data.Config.ReportOptions.ShowImports
        },
        "shouldShowInternalFuncs": func(data TemplateData) bool {
            return data.Config.ReportOptions.ShowInternalFuncs
        },
        "shouldShowTests": func(data TemplateData) bool {
            return data.Config.ReportOptions.ShowTests
        },
        "shouldShowExamples": func(data TemplateData) bool {
            return data.Config.ReportOptions.ShowExamples
        },
    })

    template.Must(t.Parse(baseTemplate))
    return &Template{tmpl: t}
}

// Generate gera a documentação HTML
func (t *Template) Generate(data interface{}) (string, error) {
    var buf bytes.Buffer
    if err := t.tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("erro ao executar template: %w", err)
    }
    return buf.String(), nil
}

// baseTemplate é o template HTML base
const baseTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Documentação do Projeto</title>
    <style>
        body {
            font-family: system-ui, -apple-system, sans-serif;
            line-height: 1.5;
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
            color: #333;
        }
        
        h1, h2, h3, h4, h5, h6 {
            color: #2d3748;
            margin-top: 2rem;
        }
        
        details {
            margin: 1rem 0;
            padding: 0.5rem;
            border: 1px solid #e2e8f0;
            border-radius: 0.375rem;
        }
        
        summary {
            cursor: pointer;
            font-weight: 600;
            color: #4a5568;
        }
        
        .indent {
            margin-left: 2rem;
        }
        
        code {
            background: #f7fafc;
            padding: 0.2rem 0.4rem;
            border-radius: 0.25rem;
            font-family: ui-monospace, monospace;
            font-size: 0.875em;
        }
        
        .tag {
            color: #718096;
            font-size: 0.875em;
        }
        
        .doc-comment {
            color: #718096;
            font-style: italic;
            margin: 0.5rem 0;
        }
    </style>
</head>
<body>
    <h1>Documentação do Projeto</h1>

    {{if .Go}}
    <section id="go-docs">
        <h2>Documentação Go</h2>
        {{range .Go.Directories}}
        <details>
            <summary>{{.Path}}</summary>
            {{range .Files}}
            <div class="indent">
                <details>
                    <summary>{{.FileName}}</summary>
                    <div class="indent">
                        <p>Package: <code>{{.Package}}</code></p>

                        {{if and .Imports (shouldShowImports $)}}
                        <details>
                            <summary>Imports</summary>
                            <div class="indent">
                                {{range .Imports}}
                                <code>{{.}}</code><br>
                                {{end}}
                            </div>
                        </details>
                        {{end}}

                        {{if and .Interfaces (shouldShow $ "interfaces")}}
                        <details>
                            <summary>Interfaces</summary>
                            <div class="indent">
                                {{range .Interfaces}}
                                <details>
                                    <summary>{{.Name}}</summary>
                                    {{if .Doc}}<div class="doc-comment">{{.Doc}}</div>{{end}}
                                    {{range .Methods}}
                                    <div class="indent">
                                        <code>{{.Name}}{{.Sig}}</code>
                                        {{if .Doc}}<div class="doc-comment">{{.Doc}}</div>{{end}}
                                    </div>
                                    {{end}}
                                </details>
                                {{end}}
                            </div>
                        </details>
                        {{end}}

                        {{if and .Structs (shouldShow $ "structs")}}
                        <details>
                            <summary>Structs</summary>
                            <div class="indent">
                                {{range .Structs}}
                                <details>
                                    <summary>{{.Name}}</summary>
                                    {{if .Doc}}<div class="doc-comment">{{.Doc}}</div>{{end}}
                                    
                                    {{if .Fields}}
                                    <details>
                                        <summary>Fields</summary>
                                        <div class="indent">
                                            {{range .Fields}}
                                            <div>
                                                <code>{{.Name}} {{.Type}}</code>
                                                {{if .Tag}}<span class="tag">{{.Tag}}</span>{{end}}
                                                {{if .Doc}}<div class="doc-comment">{{.Doc}}</div>{{end}}
                                            </div>
                                            {{end}}
                                        </div>
                                    </details>
                                    {{end}}

                                    {{if .Methods}}
                                    <details>
                                        <summary>Methods</summary>
                                        <div class="indent">
                                            {{range .Methods}}
                                            <div>
                                                <code>{{.Name}}{{.Sig}}</code>
                                                {{if .Doc}}<div class="doc-comment">{{.Doc}}</div>{{end}}
                                            </div>
                                            {{end}}
                                        </div>
                                    </details>
                                    {{end}}
                                </details>
                                {{end}}
                            </div>
                        </details>
                        {{end}}

                        {{if and .Functions (shouldShow $ "functions")}}
                        <details>
                            <summary>Functions</summary>
                            <div class="indent">
                                {{range .Functions}}
                                <div>
                                    <code>{{.Name}}{{.Sig}}</code>
                                    {{if .Doc}}<div class="doc-comment">{{.Doc}}</div>{{end}}
                                </div>
                                {{end}}
                            </div>
                        </details>
                        {{end}}
                    </div>
                </details>
            </div>
            {{end}}
        </details>
        {{end}}
    </section>
    {{end}}

    {{if .Go}}
<section id="go-docs">
    <h2>Documentação Go</h2>

    <!-- Adiciona o diagrama Mermaid -->
    <details>
        <summary>Diagrama de Classes</summary>
        <div class="mermaid">
            {{.GoMermaid}}
        </div>
    </details>

    <!-- Resto do template continua igual -->
</section>
{{end}}

    {{if .K8s}}
    <section id="k8s-docs">
        <h2>Documentação Kubernetes</h2>
        <!-- Template para recursos Kubernetes -->
        {{range .K8s.Resources}}
        <details>
            <summary>{{.Kind}}: {{.Name}}</summary>
            <div class="indent">
                {{if .Namespace}}
                <p>Namespace: <code>{{.Namespace}}</code></p>
                {{end}}
                
                {{if .Labels}}
                <details>
                    <summary>Labels</summary>
                    <div class="indent">
                        {{range $key, $value := .Labels}}
                        <code>{{$key}}: {{$value}}</code><br>
                        {{end}}
                    </div>
                </details>
                {{end}}

                {{if .Relations}}
                <details>
                    <summary>Relations</summary>
                    <div class="indent">
                        {{range .Relations}}
                        <div>{{.FromName}} → {{.ToName}} ({{.Kind}})</div>
                        {{end}}
                    </div>
                </details>
                {{end}}
            </div>
        </details>
        {{end}}
    </section>
    {{end}}
    <script src="https://cdnjs.cloudflare.com/ajax/libs/mermaid/10.6.0/mermaid.min.js"></script>
<script>
    mermaid.initialize({ startOnLoad: true });
</script>
</body>
</html>`