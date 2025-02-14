package godoc

// ProjectDoc representa a documentação completa do projeto
type ProjectDoc struct {
    Directories []DirectoryDoc `json:"directories" yaml:"directories"`
}

// DirectoryDoc representa a documentação de um diretório
type DirectoryDoc struct {
    Path  string    `json:"path"  yaml:"path"`
    Files []FileDoc `json:"files" yaml:"files"`
}

// FileDoc representa a documentação de um arquivo
type FileDoc struct {
    FileName   string      `json:"file_name"  yaml:"file_name"`
    Package    string      `json:"package"    yaml:"package"`
    Imports    []string    `json:"imports"    yaml:"imports"`
    Interfaces []Interface `json:"interfaces" yaml:"interfaces"`
    Structs    []Struct    `json:"structs"    yaml:"structs"`
    Constants  []ConstVar  `json:"constants"  yaml:"constants"`
    Variables  []ConstVar  `json:"variables"  yaml:"variables"`
    Functions  []FuncInfo  `json:"functions"  yaml:"functions"`
}

// Interface representa uma interface Go
type Interface struct {
    Name    string   `json:"name"    yaml:"name"`
    Doc     string   `json:"doc"     yaml:"doc"`
    File    string   `json:"file"    yaml:"file"`
    Line    int      `json:"line"    yaml:"line"`
    Methods []Method `json:"methods" yaml:"methods"`
}

// Method representa um método de interface
type Method struct {
    Name string `json:"name"   yaml:"name"`
    Doc  string `json:"doc"    yaml:"doc"`
    Sig  string `json:"sig"    yaml:"sig"`
    File string `json:"file"   yaml:"file"`
    Line int    `json:"line"   yaml:"line"`
}

// StructField representa um campo de struct
type StructField struct {
    Name string `json:"name" yaml:"name"`
    Type string `json:"type" yaml:"type"`
    Tag  string `json:"tag"  yaml:"tag"`
    Doc  string `json:"doc"  yaml:"doc"`
}

// Struct representa uma struct Go
type Struct struct {
    Name    string        `json:"name"         yaml:"name"`
    Doc     string        `json:"doc"          yaml:"doc"`
    File    string        `json:"file"         yaml:"file"`
    Line    int           `json:"line"         yaml:"line"`
    Fields  []StructField `json:"fields"       yaml:"fields"`
    Methods []MethodInfo  `json:"methods"      yaml:"methods"`
}

// MethodInfo representa um método de struct
type MethodInfo struct {
    Name string `json:"name" yaml:"name"`
    Doc  string `json:"doc"  yaml:"doc"`
    Sig  string `json:"sig"  yaml:"sig"`
    File string `json:"file" yaml:"file"`
    Line int    `json:"line" yaml:"line"`
}

// ConstVar representa uma constante ou variável
type ConstVar struct {
    Name string `json:"name" yaml:"name"`
    Doc  string `json:"doc"  yaml:"doc"`
    File string `json:"file" yaml:"file"`
    Line int    `json:"line" yaml:"line"`
    Type string `json:"type" yaml:"type"`
}

// FuncInfo representa uma função
type FuncInfo struct {
    Name string `json:"name" yaml:"name"`
    Doc  string `json:"doc"  yaml:"doc"`
    File string `json:"file" yaml:"file"`
    Line int    `json:"line" yaml:"line"`
    Sig  string `json:"sig"  yaml:"sig"`
}

// DirNode representa um nó na árvore de diretórios
type DirNode struct {
    Name     string
    Children []*DirNode
    Files    []string
}