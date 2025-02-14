package godoc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgardnogueira/aimap/internal/config"
)

// Generator é responsável por gerar a documentação em diferentes formatos
type Generator struct {
    projectDoc *ProjectDoc
    basePath   string
    config     config.GolangConfig
}

// NewGenerator cria um novo gerador de documentação
func NewGenerator(doc *ProjectDoc, basePath string, cfg config.GolangConfig) *Generator {
    return &Generator{
        projectDoc: doc,
        basePath:   basePath,
        config:     cfg,
    }
}

// GenerateDirectoryTree gera a árvore de diretórios do projeto
func (g *Generator) GenerateDirectoryTree() *DirNode {
    root := &DirNode{
        Name: filepath.Base(g.basePath),
    }

    for _, dir := range g.projectDoc.Directories {
        relPath, _ := filepath.Rel(g.basePath, dir.Path)
        parts := strings.Split(relPath, string(os.PathSeparator))
        current := root

        for _, p := range parts {
            if p == "." || p == "" {
                continue
            }
            current = g.findOrCreateChild(current, p)
        }

        for _, f := range dir.Files {
            fileName := filepath.Base(f.FileName)
            current.Files = append(current.Files, fileName)
        }
    }

    return root
}

// findOrCreateChild encontra ou cria um nó filho na árvore
func (g *Generator) findOrCreateChild(node *DirNode, name string) *DirNode {
    for _, ch := range node.Children {
        if ch.Name == name {
            return ch
        }
    }
    newChild := &DirNode{Name: name}
    node.Children = append(node.Children, newChild)
    return newChild
}

// writeTreeMarkdown escreve a árvore de diretórios em formato Markdown
func (g *Generator) writeTreeMarkdown(sb *strings.Builder, node *DirNode, prefix string, isLast bool) {
    if prefix == "" {
        sb.WriteString(node.Name + "\n")
    } else {
        if isLast {
            sb.WriteString(prefix + "└── " + node.Name + "\n")
        } else {
            sb.WriteString(prefix + "├── " + node.Name + "\n")
        }
    }

    newPrefix := prefix
    if prefix == "" {
        newPrefix = "    "
    } else {
        if isLast {
            newPrefix = prefix + "    "
        } else {
            newPrefix = prefix + "│   "
        }
    }

    // Escreve arquivos
    for i, f := range node.Files {
        isLastFile := (i == len(node.Files)-1) && len(node.Children) == 0
        if isLastFile {
            sb.WriteString(newPrefix + "└── " + f + "\n")
        } else {
            sb.WriteString(newPrefix + "├── " + f + "\n")
        }
    }

    // Escreve subdiretórios
    for i, child := range node.Children {
        g.writeTreeMarkdown(sb, child, newPrefix, i == len(node.Children)-1)
    }
}

 func (g *Generator) shouldIncludeContent(contentType string) bool {
    switch g.config.ReportLevel {
    case "short":
        // No modo curto, inclui apenas informações essenciais
        return contentType == "structs" || contentType == "interfaces"
    
    case "standard":
        // No modo padrão, inclui a maioria das informações, mas não tudo
        if contentType == "imports" {
            return g.config.ReportOptions.ShowImports
        }
        if contentType == "internal_funcs" {
            return g.config.ReportOptions.ShowInternalFuncs
        }
        if contentType == "tests" {
            return g.config.ReportOptions.ShowTests
        }
        return true
    
    case "complete":
        // No modo completo, usa as opções específicas
        switch contentType {
        case "imports":
            return g.config.ReportOptions.ShowImports
        case "internal_funcs":
            return g.config.ReportOptions.ShowInternalFuncs
        case "tests":
            return g.config.ReportOptions.ShowTests
        case "examples":
            return g.config.ReportOptions.ShowExamples
        default:
            return true
        }
    
    default:
        // Por padrão, inclui tudo
        return true
    }
}

// GenerateMarkdown gera a documentação em formato Markdown
func (g *Generator) GenerateMarkdown() string {
    var sb strings.Builder

    sb.WriteString("# Documentação do Projeto\n\n")
    
    // Estrutura do projeto (sempre incluída)
    sb.WriteString("## Estrutura do Projeto\n\n")
    sb.WriteString("```\n")
    g.writeTreeMarkdown(&sb, g.GenerateDirectoryTree(), "", true)
    sb.WriteString("```\n\n")

    for _, dir := range g.projectDoc.Directories {
        relPath, _ := filepath.Rel(g.basePath, dir.Path)
        sb.WriteString(fmt.Sprintf("## Pacote: %s\n\n", relPath))

        for _, file := range dir.Files {
            // Pula arquivos de teste se não estiver configurado para mostrá-los
            if strings.HasSuffix(file.FileName, "_test.go") && !g.shouldIncludeContent("tests") {
                continue
            }

            sb.WriteString(fmt.Sprintf("### Arquivo: %s\n\n", filepath.Base(file.FileName)))
            
            // Documentar imports
            if len(file.Imports) > 0 && g.shouldIncludeContent("imports") {
                sb.WriteString("#### Imports\n\n")
                for _, imp := range file.Imports {
                    sb.WriteString(fmt.Sprintf("- `%s`\n", imp))
                }
                sb.WriteString("\n")
            }

            // Documentar interfaces
            if len(file.Interfaces) > 0 {
                sb.WriteString("#### Interfaces\n\n")
                for _, iface := range file.Interfaces {
                    g.writeInterfaceMarkdown(&sb, iface)
                }
            }

            // Documentar structs
            if len(file.Structs) > 0 {
                sb.WriteString("#### Structs\n\n")
                for _, str := range file.Structs {
                    g.writeStructMarkdown(&sb, str)
                }
            }

            // Documentar constantes
            if len(file.Constants) > 0 {
                sb.WriteString("#### Constantes\n\n")
                for _, c := range file.Constants {
                    g.writeConstVarMarkdown(&sb, c)
                }
            }

            // Documentar variáveis
            if len(file.Variables) > 0 {
                sb.WriteString("#### Variáveis\n\n")
                for _, v := range file.Variables {
                    g.writeConstVarMarkdown(&sb, v)
                }
            }

            // Documentar funções
            if len(file.Functions) > 0 {
                sb.WriteString("#### Funções\n\n")
                for _, fn := range file.Functions {
                    // Pula funções internas se não estiver configurado para mostrá-las
                    if g.isInternalFunc(fn.Name) && !g.shouldIncludeContent("internal_funcs") {
                        continue
                    }
                    g.writeFunctionMarkdown(&sb, fn)
                }
            }

            sb.WriteString("\n---\n\n")
        }
    }

    return sb.String()
}
// isInternalFunc verifica se uma função é interna (começa com letra minúscula)
func (g *Generator) isInternalFunc(name string) bool {
    if len(name) == 0 {
        return false
    }
    firstChar := name[0]
    return firstChar >= 'a' && firstChar <= 'z'
}

// writeInterfaceMarkdown documenta uma interface em Markdown
func (g *Generator) writeInterfaceMarkdown(sb *strings.Builder, iface Interface) {
    sb.WriteString(fmt.Sprintf("##### Interface `%s`\n\n", iface.Name))
    if iface.Doc != "" {
        sb.WriteString(strings.TrimSpace(iface.Doc) + "\n\n")
    }
    
    if len(iface.Methods) > 0 {
        sb.WriteString("Métodos:\n\n")
        for _, m := range iface.Methods {
            // Pula métodos internos se não estiver configurado para mostrá-los
            if g.isInternalFunc(m.Name) && !g.shouldIncludeContent("internal_funcs") {
                continue
            }
            sb.WriteString(fmt.Sprintf("- `%s%s`\n", m.Name, m.Sig))
            if m.Doc != "" {
                sb.WriteString("  - " + strings.TrimSpace(m.Doc) + "\n")
            }
        }
        sb.WriteString("\n")
    }
}

// writeStructMarkdown documenta uma struct em Markdown
func (g *Generator) writeStructMarkdown(sb *strings.Builder, str Struct) {
    sb.WriteString(fmt.Sprintf("##### Struct `%s`\n\n", str.Name))
    if str.Doc != "" {
        sb.WriteString(strings.TrimSpace(str.Doc) + "\n\n")
    }

    if len(str.Fields) > 0 {
        sb.WriteString("Campos:\n\n")
        for _, f := range str.Fields {
            if g.isInternalFunc(f.Name) && !g.shouldIncludeContent("internal_funcs") {
                continue
            }
            sb.WriteString(fmt.Sprintf("- `%s %s`", f.Name, f.Type))
            if f.Tag != "" {
                sb.WriteString(fmt.Sprintf(" `%s`", f.Tag))
            }
            sb.WriteString("\n")
            if f.Doc != "" {
                sb.WriteString("  - " + strings.TrimSpace(f.Doc) + "\n")
            }
        }
        sb.WriteString("\n")
    }

    if len(str.Methods) > 0 {
        sb.WriteString("Métodos:\n\n")
        for _, m := range str.Methods {
            if g.isInternalFunc(m.Name) && !g.shouldIncludeContent("internal_funcs") {
                continue
            }
            sb.WriteString(fmt.Sprintf("- `%s%s`\n", m.Name, m.Sig))
            if m.Doc != "" {
                sb.WriteString("  - " + strings.TrimSpace(m.Doc) + "\n")
            }
        }
        sb.WriteString("\n")
    }
}





// writeConstVarMarkdown documenta uma constante ou variável em Markdown
func (g *Generator) writeConstVarMarkdown(sb *strings.Builder, cv ConstVar) {
    sb.WriteString(fmt.Sprintf("- `%s %s`\n", cv.Name, cv.Type))
    if cv.Doc != "" {
        sb.WriteString("  - " + strings.TrimSpace(cv.Doc) + "\n")
    }
}

// writeFunctionMarkdown documenta uma função em Markdown
func (g *Generator) writeFunctionMarkdown(sb *strings.Builder, fn FuncInfo) {
    sb.WriteString(fmt.Sprintf("##### `%s%s`\n\n", fn.Name, fn.Sig))
    if fn.Doc != "" {
        sb.WriteString(strings.TrimSpace(fn.Doc) + "\n\n")
    }
}

// GenerateHTML gera a documentação em formato HTML
func (g *Generator) GenerateHTML() string {
    // Implemente a geração de HTML usando templates
    // Você pode mover a implementação existente do generateHTMLDoc aqui
    return "" // TODO: implementar
}

// GenerateJSON gera a documentação em formato JSON
func (g *Generator) GenerateJSON() interface{} {
    // O ProjectDoc já está estruturado para ser serializado como JSON
    return g.projectDoc
}

// GenerateMermaidDiagram gera um diagrama Mermaid da estrutura do código
func (g *Generator) GenerateMermaidDiagram() string {
    var sb strings.Builder

    sb.WriteString("classDiagram\n")

    // Mapeia structs por pacote
    processed := make(map[string]bool)
    relationships := make(map[string]bool)

    // Processa todos os arquivos
    for _, dir := range g.projectDoc.Directories {
        for _, file := range dir.Files {
            for _, str := range file.Structs {
                className := fmt.Sprintf("%s_%s", file.Package, str.Name)
                if processed[className] {
                    continue
                }
                processed[className] = true

                // Define a classe
                sb.WriteString(fmt.Sprintf("    class %s {\n", className))
                
                // Campos
                for _, field := range str.Fields {
                    fieldType := cleanTypeForDisplay(field.Type)
                    sb.WriteString(fmt.Sprintf("        +%s %s\n", field.Name, fieldType))

                    // Adiciona relacionamento se for um tipo definido
                    if relatedType := getRelatedType(field.Type); relatedType != "" {
                        relationship := fmt.Sprintf("    %s --> %s\n", className, relatedType)
                        relationships[relationship] = true
                    }
                }

                // Métodos principais (limitado a 5)
                methodCount := 0
                for _, method := range str.Methods {
                    if methodCount >= 5 {
                        break
                    }
                    sb.WriteString(fmt.Sprintf("        +%s()\n", method.Name))
                    methodCount++
                }
                sb.WriteString("    }\n")
            }
        }
    }

    // Adiciona relacionamentos
    for rel := range relationships {
        sb.WriteString(rel)
    }

    return sb.String()
}

// cleanTypeForDisplay limpa o tipo para exibição no diagrama
func cleanTypeForDisplay(t string) string {
    // Remove ponteiros
    t = strings.TrimPrefix(t, "*")

    // Simplifica arrays e slices
    if strings.HasPrefix(t, "[]") {
        return "Array"
    }

    // Simplifica maps
    if strings.HasPrefix(t, "map[") {
        return "Map"
    }

    // Simplifica channels
    if strings.Contains(t, "chan") {
        return "Channel"
    }

    // Simplifica tipos do pacote regexp
    if strings.Contains(t, "regexp.Regexp") {
        return "RegExp"
    }

    // Remove tipos genéricos
    if idx := strings.Index(t, "<"); idx != -1 {
        t = t[:idx]
    }

    // Preserva apenas o nome do tipo após o ponto
    if idx := strings.LastIndex(t, "."); idx != -1 {
        t = t[idx+1:]
    }

    return t
}

// getRelatedType retorna o tipo relacionado formatado para o diagrama
func getRelatedType(t string) string {
    // Remove ponteiros e arrays
    t = strings.TrimPrefix(t, "*")
    t = strings.TrimPrefix(t, "[]")

    // Ignora tipos básicos e especiais
    if isBasicType(t) || isSpecialType(t) {
        return ""
    }

    // Se for um tipo do pacote atual, usa apenas o nome
    if !strings.Contains(t, ".") {
        return t
    }

    // Para tipos de outros pacotes, formata como pkg_type
    parts := strings.Split(t, ".")
    if len(parts) == 2 {
        return fmt.Sprintf("%s_%s", parts[0], parts[1])
    }

    return ""
}

// isBasicType verifica se é um tipo básico do Go
func isBasicType(t string) bool {
    basics := map[string]bool{
        "string": true, "int": true, "bool": true,
        "int8": true, "int16": true, "int32": true, "int64": true,
        "uint8": true, "uint16": true, "uint32": true, "uint64": true,
        "float32": true, "float64": true, "complex64": true, "complex128": true,
        "byte": true, "rune": true, "error": true, "interface{}": true,
    }
    return basics[t]
}

// isSpecialType verifica se é um tipo especial que deve ser ignorado nas relações
func isSpecialType(t string) bool {
    return strings.Contains(t, "regexp.Regexp") ||
           strings.Contains(t, "template.Template") ||
           strings.Contains(t, "ast.") ||  // Tipos do pacote ast
           strings.Contains(t, "time.") || // Tipos do pacote time
           strings.HasPrefix(t, "map[") ||
           strings.Contains(t, "chan")
}
// simplifyType retorna uma versão simplificada do tipo
func simplifyType(t string) string {
    t = strings.TrimPrefix(t, "*")
    if strings.HasPrefix(t, "[]") {
        return "Array"
    }
    if strings.HasPrefix(t, "map[") {
        return "Map"
    }
    if idx := strings.Index(t, "."); idx != -1 {
        return t[idx+1:]
    }
    return t
}

// isStructType verifica se é um tipo struct
func isStructType(t string) bool {
    t = strings.TrimPrefix(t, "*")
    return strings.Contains(t, ".")
}

// getRelatedClass retorna o nome da classe relacionada
func getRelatedClass(t string) string {
    t = strings.TrimPrefix(t, "*")
    if idx := strings.Index(t, "."); idx != -1 {
        pkg := t[:idx]
        name := t[idx+1:]
        return fmt.Sprintf("%s_%s", pkg, name)
    }
    return t
}
// cleanName limpa o nome para uso no diagrama Mermaid
func cleanName(name string) string {
    return strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"), "*", "")
}

// cleanSignature limpa a assinatura do método para uso no diagrama
func cleanSignature(sig string) string {
    // Remove caracteres problemáticos e simplifica a assinatura
    sig = strings.ReplaceAll(sig, "<", "")
    sig = strings.ReplaceAll(sig, ">", "")
    sig = strings.ReplaceAll(sig, "&", "")
    return strings.ReplaceAll(sig, "|", "")
}

// cleanType limpa o tipo para uso no diagrama
func cleanType(typeName string) string {
    // Simplifica tipos genéricos e ponteiros
    typeName = strings.ReplaceAll(typeName, "*", "")
    if idx := strings.Index(typeName, "<"); idx != -1 {
        typeName = typeName[:idx]
    }
    return typeName
}

// isStructReference verifica se um tipo é uma referência a outra struct
func isStructReference(typeName string) bool {
    // Remove ponteiros e verifica se não é um tipo básico
    typeName = strings.TrimPrefix(typeName, "*")
    basicTypes := map[string]bool{
        "string": true, "int": true, "bool": true, "interface{}": true,
        "error": true, "byte": true, "rune": true,
    }
    return !basicTypes[typeName] && !strings.Contains(typeName, "[]") && 
           !strings.Contains(typeName, "map[") && !strings.Contains(typeName, "chan")
}

// getBaseType obtém o tipo base de um tipo
func getBaseType(typeName string) string {
    // Remove ponteiros, arrays, etc.
    typeName = strings.TrimPrefix(typeName, "*")
    typeName = strings.TrimPrefix(typeName, "[]")
    if idx := strings.Index(typeName, "<"); idx != -1 {
        typeName = typeName[:idx]
    }
    return typeName
}
// isInternalField verifica se um campo/método é interno (começa com letra minúscula)
func (g *Generator) isInternalField(name string) bool {
    if len(name) == 0 {
        return false
    }
    firstChar := name[0]
    return firstChar >= 'a' && firstChar <= 'z'
}

// structImplementsInterface verifica se uma struct implementa uma interface
func (g *Generator) structImplementsInterface(str Struct, iface Interface) bool {
    // Cria um mapa dos métodos da struct
    structMethods := make(map[string]string)
    for _, method := range str.Methods {
        structMethods[method.Name] = method.Sig
    }

    // Verifica se todos os métodos da interface estão presentes na struct
    for _, ifaceMethod := range iface.Methods {
        structSig, ok := structMethods[ifaceMethod.Name]
        if !ok || structSig != ifaceMethod.Sig {
            return false
        }
    }

    return true
}