package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Template representa um template Markdown
type Template struct {
    tmpl *template.Template
}
type TemplateData struct {
    Go        interface{}
    K8s       interface{}
    Config    interface{}
    GoMermaid string
}

// NewTemplate cria um novo template Markdown
func NewTemplate() *Template {
    t := template.New("documentation").Funcs(template.FuncMap{
        "indent": func(spaces int, v string) string {
            pad := strings.Repeat(" ", spaces)
            return pad + strings.Replace(v, "\n", "\n"+pad, -1)
        },
        "codeBlock": func(lang, code string) string {
            return "```" + lang + "\n" + code + "\n```"
        },
    })

    template.Must(t.Parse(baseTemplate))
    return &Template{tmpl: t}
}

// Generate gera a documentação Markdown
func (t *Template) Generate(data interface{}) (string, error) {
    var buf bytes.Buffer
    if err := t.tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("erro ao executar template: %w", err)
    }
    return buf.String(), nil
}

// baseTemplate é o template Markdown base
const baseTemplate = `# Documentação do Projeto

{{if .Go}}
## Documentação Go

### Diagrama de Classes
` + "```mermaid" + `
{{.GoMermaid}}
` + "```" + `

### Estrutura do Projeto
` + "```" + `
{{range .Go.Directories}}
{{.Path}}
{{range .Files}}  ├── {{.FileName}}
{{end}}
{{end}}
` + "```" + `

{{range .Go.Directories}}
### 📁 {{.Path}}

{{range .Files}}
#### 📄 {{.FileName}}

**Package:** ` + "`{{.Package}}`" + `

{{if .Imports}}
##### Imports

{{range .Imports}}
- ` + "`{{.}}`" + `
{{end}}
{{end}}

{{if .Interfaces}}
##### Interfaces

{{range .Interfaces}}
###### Interface ` + "`{{.Name}}`" + `

{{if .Doc}}
{{.Doc}}
{{end}}

{{range .Methods}}
- ` + "`{{.Name}}{{.Sig}}`" + `
{{if .Doc}}  - {{.Doc}}{{end}}
{{end}}

{{end}}
{{end}}

{{if .Structs}}
##### Structs

{{range .Structs}}
###### Struct ` + "`{{.Name}}`" + `

{{if .Doc}}
{{.Doc}}
{{end}}

{{if .Fields}}
**Fields:**

{{range .Fields}}
- ` + "`{{.Name}} {{.Type}}`" + ` {{if .Tag}}` + "`{{.Tag}}`" + `{{end}}
{{if .Doc}}  - {{.Doc}}{{end}}
{{end}}
{{end}}

{{if .Methods}}
**Methods:**

{{range .Methods}}
- ` + "`{{.Name}}{{.Sig}}`" + `
{{if .Doc}}  - {{.Doc}}{{end}}
{{end}}
{{end}}

{{end}}
{{end}}

{{if .Functions}}
##### Functions

{{range .Functions}}
###### ` + "`{{.Name}}{{.Sig}}`" + `

{{if .Doc}}
{{.Doc}}
{{end}}
{{end}}
{{end}}

---
{{end}}
{{end}}
{{end}}

{{if .K8s}}
## Documentação Kubernetes

{{range .K8s.Resources}}
### {{.Kind}}: {{.Name}}

{{if .Namespace}}
**Namespace:** ` + "`{{.Namespace}}`" + `
{{end}}

{{if .Labels}}
#### Labels

{{range $key, $value := .Labels}}
- ` + "`{{$key}}: {{$value}}`" + `
{{end}}
{{end}}

{{if .Relations}}
#### Relations

{{range .Relations}}
- {{.FromName}} → {{.ToName}} ({{.Kind}})
{{end}}
{{end}}

---
{{end}}
{{end}}

## Diagrama de Relacionamentos

` + "```mermaid" + `
graph TD
{{range .K8s.Resources}}
    {{.Name}}[{{.Kind}}: {{.Name}}]
    {{range .Relations}}
    {{.FromName}} --> {{.ToName}}
    {{end}}
{{end}}
` + "```" + `
`