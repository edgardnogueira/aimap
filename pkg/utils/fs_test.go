package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileUtils(t *testing.T) {
    // Criar FileUtils com alguns padrões de ignore
    fu, err := NewFileUtils([]string{
        ".*\\.tmp$",
        ".*\\.bak$",
        "test/ignore/.*",
    })
    if err != nil {
        t.Fatalf("Erro ao criar FileUtils: %v", err)
    }

    // Testar ShouldIgnore
    t.Run("ShouldIgnore", func(t *testing.T) {
        cases := []struct {
            path     string
            expected bool
        }{
            {"file.go", false},
            {"file.tmp", true},
            {"file.bak", true},
            {"test/ignore/file.go", true},
            {"test/valid/file.go", false},
        }

        for _, tc := range cases {
            if got := fu.ShouldIgnore(tc.path); got != tc.expected {
                t.Errorf("ShouldIgnore(%q) = %v; want %v", tc.path, got, tc.expected)
            }
        }
    })

    // Criar diretório temporário para testes
    tempDir, err := os.MkdirTemp("", "fileutils_test_*")
    if err != nil {
        t.Fatalf("Erro ao criar diretório temporário: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Testar EnsureDir
    t.Run("EnsureDir", func(t *testing.T) {
        testDir := filepath.Join(tempDir, "test", "nested")
        if err := fu.EnsureDir(testDir); err != nil {
            t.Errorf("EnsureDir(%q) falhou: %v", testDir, err)
        }
        if _, err := os.Stat(testDir); os.IsNotExist(err) {
            t.Errorf("Diretório %q não foi criado", testDir)
        }
    })

    // Testar WriteLines e ReadLines
    t.Run("WriteAndReadLines", func(t *testing.T) {
        testFile := filepath.Join(tempDir, "test.txt")
        lines := []string{"linha 1", "linha 2", "linha 3"}

        if err := fu.WriteLines(testFile, lines); err != nil {
            t.Fatalf("WriteLines falhou: %v", err)
        }

        readLines, err := fu.ReadLines(testFile)
        if err != nil {
            t.Fatalf("ReadLines falhou: %v", err)
        }

        if len(readLines) != len(lines) {
            t.Errorf("Número de linhas não corresponde: got %d, want %d", len(readLines), len(lines))
        }

        for i, line := range lines {
            if readLines[i] != line {
                t.Errorf("Linha %d não corresponde: got %q, want %q", i, readLines[i], line)
            }
        }
    })

    // Testar CopyFile
    t.Run("CopyFile", func(t *testing.T) {
        srcFile := filepath.Join(tempDir, "source.txt")
        dstFile := filepath.Join(tempDir, "dest", "copied.txt")
        content := []byte("teste de conteúdo")

        if err := os.WriteFile(srcFile, content, 0644); err != nil {
            t.Fatalf("Erro ao criar arquivo fonte: %v", err)
        }

        if err := fu.CopyFile(srcFile, dstFile); err != nil {
            t.Fatalf("CopyFile falhou: %v", err)
        }

        copiedContent, err := os.ReadFile(dstFile)
        if err != nil {
            t.Fatalf("Erro ao ler arquivo copiado: %v", err)
        }

        if string(copiedContent) != string(content) {
            t.Errorf("Conteúdo copiado não corresponde: got %q, want %q", string(copiedContent), string(content))
        }
    })

    // Testar FindFiles
    t.Run("FindFiles", func(t *testing.T) {
        // Criar alguns arquivos para teste
        files := []string{
            filepath.Join(tempDir, "test1.go"),
            filepath.Join(tempDir, "test2.go"),
            filepath.Join(tempDir, "test.txt"),
        }

        for _, f := range files {
            if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
                t.Fatalf("Erro ao criar arquivo de teste: %v", err)
            }
        }

        matches, err := fu.FindFiles([]string{tempDir}, "*.go")
        if err != nil {
            t.Fatalf("FindFiles falhou: %v", err)
        }

        if len(matches) != 2 {
            t.Errorf("Número incorreto de arquivos encontrados: got %d, want 2", len(matches))
        }
    })
}