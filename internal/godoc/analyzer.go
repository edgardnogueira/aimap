package godoc

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/edgardnogueira/aimap/internal/config"
)

// Analyzer é responsável por analisar o código Go
type Analyzer struct {
    config config.GolangConfig
    ignoreRegex []*regexp.Regexp
}

// NewAnalyzer cria um novo analisador de código Go
func NewAnalyzer(cfg config.GolangConfig) *Analyzer {
    var regexps []*regexp.Regexp
    for _, pattern := range cfg.Ignores {
        if re, err := regexp.Compile(pattern); err == nil {
            regexps = append(regexps, re)
        } else {
            slog.Warn("Padrão de ignore inválido", "pattern", pattern, "error", err)
        }
    }

    return &Analyzer{
        config: cfg,
        ignoreRegex: regexps,
    }
}

// Analyze analisa todos os diretórios configurados
func (a *Analyzer) Analyze() (*ProjectDoc, error) {
    projectDoc := &ProjectDoc{}

    for _, path := range a.config.Paths {
        if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }

            if a.shouldIgnore(path) {
                if d.IsDir() {
                    return filepath.SkipDir
                }
                return nil
            }

            if d.IsDir() {
                files, err := a.analyzeDirectory(path)
                if err != nil {
                    slog.Error("Erro ao analisar diretório", "path", path, "error", err)
                    return nil
                }
                if len(files) > 0 {
                    projectDoc.Directories = append(projectDoc.Directories, DirectoryDoc{
                        Path:  path,
                        Files: files,
                    })
                }
            }
            return nil
        }); err != nil {
            return nil, err
        }
    }

    return projectDoc, nil
}

// shouldIgnore verifica se um caminho deve ser ignorado
func (a *Analyzer) shouldIgnore(path string) bool {
    for _, re := range a.ignoreRegex {
        if re.MatchString(path) {
            return true
        }
    }
    return false
}

// analyzeDirectory analisa um diretório específico
func (a *Analyzer) analyzeDirectory(dirPath string) ([]FileDoc, error) {
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        return nil, err
    }

    var files []FileDoc
    for _, entry := range entries {
        if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".go") {
            continue
        }

        fullPath := filepath.Join(dirPath, entry.Name())
        if a.shouldIgnore(fullPath) {
            continue
        }

        fileDoc, err := a.analyzeFile(fullPath)
        if err != nil {
            slog.Error("Erro ao analisar arquivo", "path", fullPath, "error", err)
            continue
        }
        files = append(files, fileDoc)
    }

    return files, nil
}
// analyzeFile analisa um arquivo Go específico
func (a *Analyzer) analyzeFile(filePath string) (FileDoc, error) {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
    if err != nil {
        return FileDoc{}, err
    }

    fileDoc := FileDoc{
        FileName: filePath,
        Package:  node.Name.Name,
        Imports:  a.collectImports(node),
    }

    // Analisa as declarações do arquivo
    for _, decl := range node.Decls {
        switch d := decl.(type) {
        case *ast.GenDecl:
            switch d.Tok {
            case token.CONST:
                consts := a.collectConstVar(d, fset, filePath, true)
                fileDoc.Constants = append(fileDoc.Constants, consts...)
            case token.VAR:
                vars := a.collectConstVar(d, fset, filePath, false)
                fileDoc.Variables = append(fileDoc.Variables, vars...)
            case token.TYPE:
                a.collectTypes(d, fset, &fileDoc) // Corrigido aqui: removido o parâmetro filePath
            }
        case *ast.FuncDecl:
            a.collectFunction(d, fset, filePath, &fileDoc)
        }
    }

    return fileDoc, nil
}
// Aqui viriam as funções auxiliares como collectImports, collectConstVar, 
// collectTypes, collectFunction, etc., que você já tem no código original.
// Elas seriam movidas do arquivo principal para cá, mantendo a mesma lógica
// mas adaptadas para usar os métodos do Analyzer.

// collectImports coleta todos os imports do arquivo
func (a *Analyzer) collectImports(f *ast.File) []string {
    var imports []string
    for _, imp := range f.Imports {
        path := strings.Trim(imp.Path.Value, `"`)
        imports = append(imports, path)
    }
    return imports
}

// collectConstVar coleta constantes ou variáveis
func (a *Analyzer) collectConstVar(decl *ast.GenDecl, fset *token.FileSet, fileName string, isConst bool) []ConstVar {
    var result []ConstVar
    for _, spec := range decl.Specs {
        vs, ok := spec.(*ast.ValueSpec)
        if !ok {
            continue
        }
        
        for _, name := range vs.Names {
            pos := fset.Position(name.Pos())
            cv := ConstVar{
                Name: name.Name,
                Doc:  strings.TrimSpace(decl.Doc.Text()),
                File: pos.Filename,
                Line: pos.Line,
            }
            
            if vs.Type != nil {
                cv.Type = a.exprToString(vs.Type)
            }
            result = append(result, cv)
        }
    }
    return result
}

// collectTypes coleta tipos (interfaces e structs)
func (a *Analyzer) collectTypes(decl *ast.GenDecl, fset *token.FileSet, fileDoc *FileDoc) {
    for _, spec := range decl.Specs {
        typeSpec, ok := spec.(*ast.TypeSpec)
        if !ok {
            continue
        }
        
        pos := fset.Position(typeSpec.Pos())
        doc := decl.Doc.Text()
        if typeSpec.Doc != nil && typeSpec.Doc.Text() != "" {
            doc = typeSpec.Doc.Text()
        }

        switch t := typeSpec.Type.(type) {
        case *ast.InterfaceType:
            iface := Interface{
                Name: typeSpec.Name.Name,
                Doc:  strings.TrimSpace(doc),
                File: pos.Filename,
                Line: pos.Line,
            }
            
            if t.Methods != nil {
                for _, m := range t.Methods.List {
                    if len(m.Names) == 0 {
                        continue
                    }
                    for _, name := range m.Names {
                        methodPos := fset.Position(m.Pos())
                        methodDoc := ""
                        if m.Doc != nil {
                            methodDoc = m.Doc.Text()
                        }
                        if funcType, ok := m.Type.(*ast.FuncType); ok {
                            sig := a.buildFuncSig(funcType)
                            iface.Methods = append(iface.Methods, Method{
                                Name: name.Name,
                                Doc:  strings.TrimSpace(methodDoc),
                                Sig:  sig,
                                File: methodPos.Filename,
                                Line: methodPos.Line,
                            })
                        }
                    }
                }
            }
            fileDoc.Interfaces = append(fileDoc.Interfaces, iface)

        case *ast.StructType:
            str := Struct{
                Name: typeSpec.Name.Name,
                Doc:  strings.TrimSpace(doc),
                File: pos.Filename,
                Line: pos.Line,
            }
            
            if t.Fields != nil {
                for _, field := range t.Fields.List {
                    fieldDoc := ""
                    if field.Doc != nil {
                        fieldDoc = field.Doc.Text()
                    }
                    tagStr := ""
                    if field.Tag != nil {
                        tagStr = strings.Trim(field.Tag.Value, "`")
                    }
                    fieldType := a.exprToString(field.Type)
                    
                    if len(field.Names) == 0 {
                        // Campo anônimo
                        str.Fields = append(str.Fields, StructField{
                            Name: fieldType,
                            Type: fieldType,
                            Tag:  tagStr,
                            Doc:  strings.TrimSpace(fieldDoc),
                        })
                    } else {
                        for _, name := range field.Names {
                            str.Fields = append(str.Fields, StructField{
                                Name: name.Name,
                                Type: fieldType,
                                Tag:  tagStr,
                                Doc:  strings.TrimSpace(fieldDoc),
                            })
                        }
                    }
                }
            }
            fileDoc.Structs = append(fileDoc.Structs, str)
        }
    }
}

// collectFunction coleta informações de uma função
func (a *Analyzer) collectFunction(decl *ast.FuncDecl, fset *token.FileSet, fileName string, fileDoc *FileDoc) {
    pos := fset.Position(decl.Pos())
    doc := ""
    if decl.Doc != nil {
        doc = decl.Doc.Text()
    }
    
    sig := a.buildFuncSig(decl.Type)
    
    if decl.Recv == nil {
        // Função normal (não é método)
        fileDoc.Functions = append(fileDoc.Functions, FuncInfo{
            Name: decl.Name.Name,
            Doc:  strings.TrimSpace(doc),
            File: pos.Filename,
            Line: pos.Line,
            Sig:  sig,
        })
    } else {
        // É um método, adiciona ao struct correspondente
        receiverName := a.extractReceiverTypeName(decl.Recv.List[0].Type)
        if receiverName == "" {
            return
        }
        
        method := MethodInfo{
            Name: decl.Name.Name,
            Doc:  strings.TrimSpace(doc),
            File: pos.Filename,
            Line: pos.Line,
            Sig:  sig,
        }
        
        // Procura o struct correspondente
        for i := range fileDoc.Structs {
            if fileDoc.Structs[i].Name == receiverName {
                fileDoc.Structs[i].Methods = append(fileDoc.Structs[i].Methods, method)
                break
            }
        }
    }
}

// buildFuncSig constrói a assinatura de uma função
func (a *Analyzer) buildFuncSig(funcType *ast.FuncType) string {
    params := a.extractFieldList(funcType.Params)
    results := a.extractFieldList(funcType.Results)
    
    if results == "" {
        return fmt.Sprintf("(%s)", params)
    }
    return fmt.Sprintf("(%s) (%s)", params, results)
}

// extractFieldList extrai a lista de campos de uma função
func (a *Analyzer) extractFieldList(fieldList *ast.FieldList) string {
    if fieldList == nil || len(fieldList.List) == 0 {
        return ""
    }
    
    var parts []string
    for _, field := range fieldList.List {
        names := make([]string, len(field.Names))
        for i, name := range field.Names {
            names[i] = name.Name
        }
        typeExpr := a.exprToString(field.Type)
        
        if len(names) > 0 {
            parts = append(parts, fmt.Sprintf("%s %s", strings.Join(names, ", "), typeExpr))
        } else {
            parts = append(parts, typeExpr)
        }
    }
    return strings.Join(parts, ", ")
}

// extractReceiverTypeName extrai o nome do tipo receptor de um método
func (a *Analyzer) extractReceiverTypeName(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.StarExpr:
        if id, ok := t.X.(*ast.Ident); ok {
            return id.Name
        }
    }
    return ""
}

// exprToString converte uma expressão AST em string
func (a *Analyzer) exprToString(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.StarExpr:
        return "*" + a.exprToString(t.X)
    case *ast.SelectorExpr:
        return a.exprToString(t.X) + "." + t.Sel.Name
    case *ast.ArrayType:
        return "[]" + a.exprToString(t.Elt)
    case *ast.MapType:
        return fmt.Sprintf("map[%s]%s", a.exprToString(t.Key), a.exprToString(t.Value))
    case *ast.InterfaceType:
        return "interface{}"
    case *ast.Ellipsis:
        return "..." + a.exprToString(t.Elt)
    case *ast.FuncType:
        return "func" + a.buildFuncSig(t)
    case *ast.ChanType:
        switch t.Dir {
        case ast.SEND:
            return "chan<- " + a.exprToString(t.Value)
        case ast.RECV:
            return "<-chan " + a.exprToString(t.Value)
        default:
            return "chan " + a.exprToString(t.Value)
        }
    }
    return fmt.Sprintf("%T", expr)
}