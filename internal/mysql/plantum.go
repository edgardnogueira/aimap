// internal/mysql/plantuml.go
package mysql

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
	sb.WriteString("skinparam linetype ortho\n")
	sb.WriteString("skinparam roundcorner 5\n")
	sb.WriteString("skinparam shadowing false\n")
	sb.WriteString("skinparam handwritten false\n")
	sb.WriteString("skinparam class {\n")
	sb.WriteString("  BackgroundColor White\n")
	sb.WriteString("  ArrowColor Gray\n")
	sb.WriteString("  BorderColor Gray\n")
	sb.WriteString("}\n\n")
	
	// Título
	sb.WriteString(fmt.Sprintf("title Database Schema: %s\n\n", g.database.Name))

	// Gera entidades para tabelas
	for _, table := range g.database.Tables {
		g.generateTable(&sb, table)
		sb.WriteString("\n")
	}

	// Gera entidades para views
	for _, view := range g.database.Views {
		g.generateView(&sb, view)
		sb.WriteString("\n")
	}

	// Gera relacionamentos
	for _, table := range g.database.Tables {
		g.generateRelationships(&sb, table)
	}

	// Rodapé do diagrama
	sb.WriteString("\n@enduml")

	return sb.String()
}

func (g *PlantUMLGenerator) generateTable(sb *strings.Builder, table Table) {
	tableName := sanitizeTableName(table.Name)
	// Início da entidade
	sb.WriteString(fmt.Sprintf("entity \"%s\" as %s {\n", table.Name, tableName))

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
			if idx.Name != "PRIMARY" {
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
	}

	sb.WriteString("}\n")

	// Adiciona nota se houver comentário
	if table.Comment != "" {
		sb.WriteString(fmt.Sprintf("note bottom of %s\n", tableName))
		sb.WriteString(table.Comment)
		sb.WriteString("\nend note\n")
	}
}

func (g *PlantUMLGenerator) generateView(sb *strings.Builder, view View) {
	viewName := sanitizeTableName(view.Name)
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

func (g *PlantUMLGenerator) generateRelationships(sb *strings.Builder, table Table) {
	for _, fk := range table.ForeignKeys {
		// Gera a relação usando a sintaxe correta do PlantUML
		sb.WriteString(fmt.Sprintf("%s \"*\" -- \"1\" %s",
			sanitizeTableName(fk.RefTable),
			sanitizeTableName(table.Name)))

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
}

func sanitizeTableName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func formatColumnType(col Column) string {
	var parts []string
	parts = append(parts, col.Type)
	
	if col.Default != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", col.Default))
	}
	if col.Extra != "" {
		parts = append(parts, col.Extra)
	}
	
	return strings.Join(parts, " ")
}