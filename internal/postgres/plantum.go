// internal/postgres/plantuml.go
package postgres

import (
	"fmt"
	"strings"
)

type PlantUMLGenerator struct {
	database *Database
}

func NewPlantUMLGenerator(db *Database) *PlantUMLGenerator {
	return &PlantUMLGenerator{database: db}
}

func (g *PlantUMLGenerator) Generate() string {
	var sb strings.Builder

	// Cabeçalho do diagrama
	sb.WriteString("@startuml\n\n")
	sb.WriteString("!theme plain\n")
	sb.WriteString("' Configurações para reduzir dependência do Graphviz\n")
	sb.WriteString("skinparam linetype polyline\n")
	sb.WriteString("skinparam ranksep 80\n")
	sb.WriteString("skinparam nodesep 80\n")
	sb.WriteString("skinparam roundcorner 5\n")
	sb.WriteString("skinparam shadowing false\n")
	sb.WriteString("skinparam handwritten false\n")
	sb.WriteString("skinparam class {\n")
	sb.WriteString("  BackgroundColor White\n")
	sb.WriteString("  ArrowColor Gray\n")
	sb.WriteString("  BorderColor Gray\n")
	sb.WriteString("}\n")
	sb.WriteString("' Desativa o uso do Graphviz para layouts mais simples\n")
	sb.WriteString("set namespaceSeparator none\n\n")

	// Título
	sb.WriteString(fmt.Sprintf("title Database Schema: %s\n\n", g.database.Name))

	// Processa cada schema
	for _, schema := range g.database.Schemas {
		// Adiciona um pacote para o schema
		sb.WriteString(fmt.Sprintf("package \"%s\" {\n", schema.Name))

		// Gera entidades para tabelas
		for _, table := range schema.Tables {
			g.generateTable(&sb, table)
			sb.WriteString("\n")
		}

		// Gera entidades para views
		for _, view := range schema.Views {
			g.generateView(&sb, view)
			sb.WriteString("\n")
		}

		// Gera entidades para views materializadas
		for _, matview := range schema.MatViews {
			g.generateMatView(&sb, matview)
			sb.WriteString("\n")
		}

		sb.WriteString("}\n\n")
	}

	// Gera relacionamentos
	for _, schema := range g.database.Schemas {
		for _, table := range schema.Tables {
			g.generateRelationships(&sb, table)
		}
	}

	// Rodapé do diagrama
	sb.WriteString("\n@enduml")

	return sb.String()
}

func (g *PlantUMLGenerator) generateTable(sb *strings.Builder, table Table) {
	entityName := sanitizeTableName(fmt.Sprintf("%s.%s", table.Schema, table.Name))
	
	// Início da entidade
	sb.WriteString(fmt.Sprintf("entity \"%s\" as %s {\n", table.Name, entityName))

	// Colunas primárias
	var hasPrimaryKey bool
	for _, col := range table.Columns {
		if col.PrimaryKey {
			hasPrimaryKey = true
			sb.WriteString(fmt.Sprintf("  * %s : %s", col.Name, formatColumnType(col)))
			if col.Comment != "" {
				sb.WriteString(fmt.Sprintf(" <<%s>>", col.Comment))
			}
			sb.WriteString("\n")
		}
	}

	if hasPrimaryKey {
		sb.WriteString("  --\n")
	}

	// Outras colunas
	for _, col := range table.Columns {
		if !col.PrimaryKey {
			nullable := "o"
			if !col.Nullable {
				nullable = "+"
			}
			sb.WriteString(fmt.Sprintf("  %s %s : %s", 
				nullable,
				col.Name, 
				formatColumnType(col)))
			if col.Comment != "" {
				sb.WriteString(fmt.Sprintf(" <<%s>>", col.Comment))
			}
			sb.WriteString("\n")
		}
	}

	// Índices
	if len(table.Indexes) > 0 {
		sb.WriteString("  --\n")
		sb.WriteString("  .. Indexes ..\n")
		for _, idx := range table.Indexes {
			indexType := "+"
			if idx.Unique {
				indexType = "*"
			}
			sb.WriteString(fmt.Sprintf("  %s %s(%s)", 
				indexType,
				idx.Name,
				strings.Join(idx.Columns, ", ")))
			if idx.Predicate != "" {
				sb.WriteString(fmt.Sprintf(" WHERE %s", idx.Predicate))
			}
			sb.WriteString("\n")
		}
	}

	// Se a tabela herda de outras
	if len(table.Inherits) > 0 {
		sb.WriteString("  --\n")
		sb.WriteString("  .. Inherits ..\n")
		for _, parent := range table.Inherits {
			sb.WriteString(fmt.Sprintf("  + %s\n", parent))
		}
	}

	sb.WriteString("}\n")

	// Adiciona nota se houver comentário
	if table.Comment != "" {
		sb.WriteString(fmt.Sprintf("note bottom of %s\n", entityName))
		sb.WriteString(table.Comment)
		sb.WriteString("\nend note\n")
	}
}

func (g *PlantUMLGenerator) generateView(sb *strings.Builder, view View) {
	viewName := sanitizeTableName(fmt.Sprintf("%s.%s", view.Schema, view.Name))
	
	// Views são representadas como classes estereotipadas
	sb.WriteString(fmt.Sprintf("class \"%s\" as %s <<view>> {\n", 
		view.Name, 
		viewName))

	// Colunas da view
	for _, col := range view.Columns {
		sb.WriteString(fmt.Sprintf("  %s : %s\n", 
			col.Name, 
			formatColumnType(col)))
	}

	sb.WriteString("}\n")

	// Adiciona nota com a query se houver
	if view.Query != "" {
		sb.WriteString(fmt.Sprintf("note bottom of %s\n", viewName))
		sb.WriteString("Query:\n")
		sb.WriteString(view.Query)
		sb.WriteString("\nend note\n")
	}
}

func (g *PlantUMLGenerator) generateMatView(sb *strings.Builder, matview MatView) {
	matviewName := sanitizeTableName(fmt.Sprintf("%s.%s", matview.Schema, matview.Name))
	
	// Views materializadas são representadas como classes estereotipadas
	sb.WriteString(fmt.Sprintf("class \"%s\" as %s <<materialized>> {\n", 
		matview.Name, 
		matviewName))

	// Colunas da view materializada
	for _, col := range matview.Columns {
		sb.WriteString(fmt.Sprintf("  %s : %s\n", 
			col.Name, 
			formatColumnType(col)))
	}

	// Índices
	if len(matview.Indexes) > 0 {
		sb.WriteString("  --\n")
		sb.WriteString("  .. Indexes ..\n")
		for _, idx := range matview.Indexes {
			indexType := "+"
			if idx.Unique {
				indexType = "*"
			}
			sb.WriteString(fmt.Sprintf("  %s %s(%s)\n",
				indexType,
				idx.Name,
				strings.Join(idx.Columns, ", ")))
		}
	}

	sb.WriteString("}\n")

	// Adiciona nota com informações adicionais
	var notes []string
	if matview.Query != "" {
		notes = append(notes, fmt.Sprintf("Query:\n%s", matview.Query))
	}
	if matview.Refreshed != "" {
		notes = append(notes, fmt.Sprintf("Last Refreshed: %s", matview.Refreshed))
	}
	if len(notes) > 0 {
		sb.WriteString(fmt.Sprintf("note bottom of %s\n", matviewName))
		sb.WriteString(strings.Join(notes, "\n\n"))
		sb.WriteString("\nend note\n")
	}
}

func (g *PlantUMLGenerator) generateRelationships(sb *strings.Builder, table Table) {
	tableName := sanitizeTableName(fmt.Sprintf("%s.%s", table.Schema, table.Name))
	
	// Chaves estrangeiras
	for _, fk := range table.ForeignKeys {
		refTable := sanitizeTableName(fmt.Sprintf("%s.%s", fk.RefSchema, fk.RefTable))
		
		// Gera a relação
		sb.WriteString(fmt.Sprintf("%s \"*\" -- \"1\" %s",
			tableName,
			refTable))

		// Adiciona regras ON DELETE/UPDATE se existirem
		var rules []string
		if fk.OnDelete != "" {
			rules = append(rules, fmt.Sprintf("ON DELETE %s", fk.OnDelete))
		}
		if fk.OnUpdate != "" {
			rules = append(rules, fmt.Sprintf("ON UPDATE %s", fk.OnUpdate))
		}
		if len(rules) > 0 {
			sb.WriteString(fmt.Sprintf(" : %s", strings.Join(rules, ", ")))
		}

		sb.WriteString("\n")
	}

	// Herança
	for _, parent := range table.Inherits {
		parentName := sanitizeTableName(parent)
		sb.WriteString(fmt.Sprintf("%s --|> %s : inherits\n",
			tableName,
			parentName))
	}
}

func sanitizeTableName(name string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(name, ".", "_"),
		"-", "_")
}

func formatColumnType(col Column) string {
	var parts []string
	parts = append(parts, col.Type)
	
	if col.Default != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", col.Default))
	}
	
	return strings.Join(parts, " ")
}