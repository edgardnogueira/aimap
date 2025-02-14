package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// FileUtils fornece funções utilitárias para operações com arquivos
type FileUtils struct {
    ignorePatterns []*regexp.Regexp
}

// NewFileUtils cria uma nova instância de FileUtils
func NewFileUtils(ignorePatterns []string) (*FileUtils, error) {
    var regexps []*regexp.Regexp
    for _, pattern := range ignorePatterns {
        re, err := regexp.Compile(pattern)
        if err != nil {
            return nil, fmt.Errorf("padrão de ignore inválido %q: %w", pattern, err)
        }
        regexps = append(regexps, re)
    }
    
    return &FileUtils{
        ignorePatterns: regexps,
    }, nil
}

// ShouldIgnore verifica se um caminho deve ser ignorado
func (f *FileUtils) ShouldIgnore(path string) bool {
    for _, re := range f.ignorePatterns {
        if re.MatchString(path) {
            return true
        }
    }
    return false
}

// WalkMatch percorre um diretório procurando arquivos que correspondam ao padrão
func (f *FileUtils) WalkMatch(root, pattern string) ([]string, error) {
    var matches []string
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if f.ShouldIgnore(path) {
            if info.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }
        if info.IsDir() {
            return nil
        }
        if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
            return err
        } else if matched {
            matches = append(matches, path)
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return matches, nil
}

// EnsureDir garante que um diretório existe, criando-o se necessário
func (f *FileUtils) EnsureDir(path string) error {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return os.MkdirAll(path, 0755)
    }
    return nil
}

// ReadLines lê todas as linhas de um arquivo
func (f *FileUtils) ReadLines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

// WriteLines escreve linhas em um arquivo
func (f *FileUtils) WriteLines(path string, lines []string) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := bufio.NewWriter(file)
    for _, line := range lines {
        if _, err := writer.WriteString(line + "\n"); err != nil {
            return err
        }
    }
    return writer.Flush()
}

// CopyFile copia um arquivo de origem para destino
func (f *FileUtils) CopyFile(src, dst string) error {
    sourceFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    // Cria o diretório de destino se não existir
    if err := f.EnsureDir(filepath.Dir(dst)); err != nil {
        return err
    }

    destFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destFile.Close()

    _, err = io.Copy(destFile, sourceFile)
    return err
}

// FindFiles encontra arquivos que correspondam a um padrão em vários diretórios
func (f *FileUtils) FindFiles(paths []string, pattern string) ([]string, error) {
    var allMatches []string
    for _, path := range paths {
        matches, err := f.WalkMatch(path, pattern)
        if err != nil {
            return nil, fmt.Errorf("erro ao procurar em %s: %w", path, err)
        }
        allMatches = append(allMatches, matches...)
    }
    return allMatches, nil
}

// IsDirectory verifica se um caminho é um diretório
func (f *FileUtils) IsDirectory(path string) (bool, error) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        return false, err
    }
    return fileInfo.IsDir(), nil
}

// GetFileSize retorna o tamanho de um arquivo
func (f *FileUtils) GetFileSize(path string) (int64, error) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        return 0, err
    }
    return fileInfo.Size(), nil
}

// ReadFile lê todo o conteúdo de um arquivo
func (f *FileUtils) ReadFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}

// WriteFile escreve dados em um arquivo
func (f *FileUtils) WriteFile(path string, data []byte) error {
    return os.WriteFile(path, data, 0644)
}

// DeleteFile exclui um arquivo ou diretório vazio
func (f *FileUtils) DeleteFile(path string) error {
    return os.Remove(path)
}

// DeleteDirectory exclui um diretório e todo seu conteúdo
func (f *FileUtils) DeleteDirectory(path string) error {
    return os.RemoveAll(path)
}

// GetModTime retorna a data de modificação de um arquivo
func (f *FileUtils) GetModTime(path string) (int64, error) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        return 0, err
    }
    return fileInfo.ModTime().Unix(), nil
}

// CreateTempFile cria um arquivo temporário
func (f *FileUtils) CreateTempFile(dir, pattern string) (*os.File, error) {
    return os.CreateTemp(dir, pattern)
}

// CreateTempDir cria um diretório temporário
func (f *FileUtils) CreateTempDir(dir, pattern string) (string, error) {
    return os.MkdirTemp(dir, pattern)
}