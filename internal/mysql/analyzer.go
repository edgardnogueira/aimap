// internal/mysql/analyzer.go
package mysql

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Analyzer struct {
	db *sql.DB
}

func NewAnalyzer(dsn string) (*Analyzer, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao MySQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao pingar MySQL: %w", err)
	}

	return &Analyzer{db: db}, nil
}

func (a *Analyzer) Close() error {
	return a.db.Close()
}

func (a *Analyzer) Analyze(dbName string) (*Database, error) {
	database := &Database{
		Name: dbName,
	}

	// Analisa tabelas
	tables, err := a.getTables(dbName)
	if err != nil {
		return nil, err
	}
	database.Tables = tables

	// Analisa views
	views, err := a.getViews(dbName)
	if err != nil {
		return nil, err
	}
	database.Views = views

	return database, nil
}

func (a *Analyzer) getTables(dbName string) ([]Table, error) {
	query := `
		SELECT 
			TABLE_NAME
		FROM 
			information_schema.TABLES 
		WHERE 
			TABLE_SCHEMA = ? 
			AND TABLE_TYPE = 'BASE TABLE'`

	rows, err := a.db.Query(query, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.Name); err != nil {
			return nil, err
		}

		// Obtém colunas
		columns, err := a.getColumns(dbName, table.Name)
		if err != nil {
			return nil, err
		}
		table.Columns = columns

		// Obtém índices
		indexes, err := a.getIndexes(dbName, table.Name)
		if err != nil {
			return nil, err
		}
		table.Indexes = indexes

		// Obtém chaves estrangeiras
		fks, err := a.getForeignKeys(dbName, table.Name)
		if err != nil {
			return nil, err
		}
		table.ForeignKeys = fks

		tables = append(tables, table)
	}

	return tables, nil
}

func (a *Analyzer) getColumns(dbName, tableName string) ([]Column, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_KEY,
			COLUMN_DEFAULT,
			EXTRA
		FROM 
			information_schema.COLUMNS
		WHERE 
			TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
		ORDER BY 
			ORDINAL_POSITION`

	rows, err := a.db.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var isNullable, columnKey string
		var defaultValue, extra sql.NullString

		err := rows.Scan(
			&col.Name,
			&col.Type,
			&isNullable,
			&columnKey,
			&defaultValue,
			&extra,
		)
		if err != nil {
			return nil, err
		}

		col.Nullable = isNullable == "YES"
		col.PrimaryKey = columnKey == "PRI"
		if defaultValue.Valid {
			col.Default = defaultValue.String
		}
		if extra.Valid {
			col.Extra = extra.String
		}

		columns = append(columns, col)
	}

	return columns, nil
}

func (a *Analyzer) getIndexes(dbName, tableName string) ([]Index, error) {
	query := `
		SELECT 
			INDEX_NAME,
			COLUMN_NAME,
			NON_UNIQUE,
			INDEX_TYPE
		FROM 
			information_schema.STATISTICS
		WHERE 
			TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
		ORDER BY 
			INDEX_NAME, SEQ_IN_INDEX`

	rows, err := a.db.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexMap := make(map[string]*Index)
	for rows.Next() {
		var indexName, columnName, indexType string
		var nonUnique bool
		if err := rows.Scan(&indexName, &columnName, &nonUnique, &indexType); err != nil {
			return nil, err
		}

		idx, exists := indexMap[indexName]
		if !exists {
			idx = &Index{
				Name:    indexName,
				Type:    indexType,
				Unique:  !nonUnique,
				Columns: make([]string, 0),
			}
			indexMap[indexName] = idx
		}
		idx.Columns = append(idx.Columns, columnName)
	}

	var indexes []Index
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}

	return indexes, nil
}

func (a *Analyzer) getForeignKeys(dbName, tableName string) ([]ForeignKey, error) {
	query := `
		SELECT
			CONSTRAINT_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME
		FROM
			information_schema.KEY_COLUMN_USAGE
		WHERE
			TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
			AND REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY
			CONSTRAINT_NAME, ORDINAL_POSITION`

	rows, err := a.db.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fkMap := make(map[string]*ForeignKey)
	for rows.Next() {
		var constraintName, columnName, refTable, refColumn string
		if err := rows.Scan(&constraintName, &columnName, &refTable, &refColumn); err != nil {
			return nil, err
		}

		fk, exists := fkMap[constraintName]
		if !exists {
			fk = &ForeignKey{
				Name:       constraintName,
				Columns:    make([]string, 0),
				RefTable:   refTable,
				RefColumns: make([]string, 0),
			}
			fkMap[constraintName] = fk
		}
		fk.Columns = append(fk.Columns, columnName)
		fk.RefColumns = append(fk.RefColumns, refColumn)
	}

	var fks []ForeignKey
	for _, fk := range fkMap {
		// Obtém informações adicionais da chave estrangeira
		if err := a.getForeignKeyRules(dbName, tableName, fk); err != nil {
			slog.Warn("Erro ao obter regras da chave estrangeira", 
				"table", tableName,
				"fk", fk.Name,
				"error", err)
		}
		fks = append(fks, *fk)
	}

	return fks, nil
}

func (a *Analyzer) getForeignKeyRules(dbName, tableName string, fk *ForeignKey) error {
	query := `
		SELECT
			DELETE_RULE,
			UPDATE_RULE
		FROM
			information_schema.REFERENTIAL_CONSTRAINTS
		WHERE
			CONSTRAINT_SCHEMA = ?
			AND TABLE_NAME = ?
			AND CONSTRAINT_NAME = ?`

	var deleteRule, updateRule string
	err := a.db.QueryRow(query, dbName, tableName, fk.Name).Scan(&deleteRule, &updateRule)
	if err != nil {
		return err
	}

	fk.OnDelete = strings.ToUpper(deleteRule)
	fk.OnUpdate = strings.ToUpper(updateRule)
	return nil
}

func (a *Analyzer) getViews(dbName string) ([]View, error) {
	query := `
		SELECT 
			TABLE_NAME,
			VIEW_DEFINITION
		FROM 
			information_schema.VIEWS
		WHERE 
			TABLE_SCHEMA = ?`

	rows, err := a.db.Query(query, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []View
	for rows.Next() {
		var view View
		if err := rows.Scan(&view.Name, &view.Query); err != nil {
			return nil, err
		}

		// Obtém colunas da view
		columns, err := a.getColumns(dbName, view.Name)
		if err != nil {
			return nil, err
		}
		view.Columns = columns

		views = append(views, view)
	}

	return views, nil
}