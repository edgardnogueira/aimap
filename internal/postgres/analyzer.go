// internal/postgres/analyzer.go
package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lib/pq"
)

type Analyzer struct {
	db *sql.DB
}

func NewAnalyzer(connStr string) (*Analyzer, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao pingar PostgreSQL: %w", err)
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

	// Obtém todos os schemas
	schemas, err := a.getSchemas()
	if err != nil {
		return nil, err
	}

	// Para cada schema, analisa seus objetos
	for _, schemaName := range schemas {
		schema, err := a.analyzeSchema(schemaName)
		if err != nil {
			slog.Error("Erro ao analisar schema",
				"schema", schemaName,
				"error", err)
			continue
		}
		database.Schemas = append(database.Schemas, *schema)
	}

	// Analisa funções de nível de banco de dados
	functions, err := a.getFunctions("")
	if err != nil {
		slog.Error("Erro ao obter funções globais", "error", err)
	} else {
		database.Functions = functions
	}

	return database, nil
}

func (a *Analyzer) getSchemas() ([]string, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema')
		AND schema_name NOT LIKE 'pg_%'
		ORDER BY schema_name`
slog.Info("query", slog.Any("query", query))
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func (a *Analyzer) analyzeSchema(schemaName string) (*Schema, error) {
	schema := &Schema{
		Name: schemaName,
	}

	// Obtém tabelas
	tables, err := a.getTables(schemaName)
	if err != nil {
		return nil, err
	}
	schema.Tables = tables

	// Obtém views
	views, err := a.getViews(schemaName)
	if err != nil {
		return nil, err
	}
	schema.Views = views

	// Obtém views materializadas
	matViews, err := a.getMatViews(schemaName)
	if err != nil {
		return nil, err
	}
	schema.MatViews = matViews

	// Obtém funções
	functions, err := a.getFunctions(schemaName)
	if err != nil {
		return nil, err
	}
	schema.Functions = functions

	return schema, nil
}
func (a *Analyzer) getTables(schemaName string) ([]Table, error) {
	query := `
		SELECT 
			c.table_name,
			pg_get_userbyid(t.relowner) as owner,
			obj_description(t.oid) as comment,
			array_remove(array_agg(i.inhparent::regclass::text), NULL) as inherits
		FROM information_schema.tables c
		JOIN pg_class t ON t.relname = c.table_name 
			AND t.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = c.table_schema)
		LEFT JOIN pg_inherits i ON i.inhrelid = t.oid
		WHERE c.table_schema = $1
		AND c.table_type = 'BASE TABLE'
		GROUP BY c.table_name, t.relowner, t.oid
		ORDER BY c.table_name`
slog.Info("query", slog.Any("query", query))
	rows, err := a.db.Query(query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		var inherits []string
		var comment sql.NullString // usando sql.NullString para tratar possíveis NULLs

		// Utilize pq.Array para fazer a conversão do array
		if err := rows.Scan(&table.Name, &table.Owner, &comment, pq.Array(&inherits)); err != nil {
			return nil, err
		}

		if comment.Valid {
			table.Comment = comment.String
		} else {
			table.Comment = ""
		}

		// Obtém colunas
		columns, err := a.getColumns(schemaName, table.Name)
		if err != nil {
			return nil, err
		}
		table.Columns = columns

		// Obtém índices
		indexes, err := a.getIndexes(schemaName, table.Name)
		if err != nil {
			return nil, err
		}
		table.Indexes = indexes

		// Obtém chaves estrangeiras
		fks, err := a.getForeignKeys(schemaName, table.Name)
		if err != nil {
			return nil, err
		}
		table.ForeignKeys = fks

		table.Inherits = inherits
		tables = append(tables, table)
	}

	return tables, nil
}

func (a *Analyzer) getColumns(schemaName, tableName string) ([]Column, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			pg_get_expr(d.adbin, d.adrelid) as default_expr,
			col_description(t.oid, c.ordinal_position) as comment,
			c.ordinal_position = ANY(
				array(
					SELECT unnest(conkey)
					FROM pg_constraint
					WHERE conrelid = t.oid AND contype = 'p'
				)
			) as is_primary,
			pg_stats.null_frac,
			pg_stats.avg_width,
			pg_stats.n_distinct
		FROM information_schema.columns c
		JOIN pg_class t ON t.relname = c.table_name 
			AND t.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = c.table_schema)
		LEFT JOIN pg_attrdef d ON d.adrelid = t.oid 
			AND d.adnum = c.ordinal_position
		LEFT JOIN pg_stats ON pg_stats.schemaname = c.table_schema
			AND pg_stats.tablename = c.table_name
			AND pg_stats.attname = c.column_name
		WHERE c.table_schema = $1
		AND c.table_name = $2
		ORDER BY c.ordinal_position`
slog.Info("query", slog.Any("query", query))
	rows, err := a.db.Query(query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var stats ColumnStats
		var isNullable, defaultValue, defaultExpr, comment sql.NullString
		var nullFrac, distinctVals sql.NullFloat64
		var avgWidth sql.NullInt64

		err := rows.Scan(
			&col.Name,
			&col.Type,
			&isNullable,
			&defaultValue,
			&defaultExpr,
			&comment,
			&col.PrimaryKey,
			&nullFrac,
			&avgWidth,
			&distinctVals,
		)
		if err != nil {
			return nil, err
		}

		col.Nullable = isNullable.String == "YES"
		
		if defaultValue.Valid {
			col.Default = defaultValue.String
		} else if defaultExpr.Valid {
			col.Default = defaultExpr.String
		}

		if comment.Valid {
			col.Comment = comment.String
		}

		// Adiciona estatísticas se disponíveis
		if nullFrac.Valid || avgWidth.Valid || distinctVals.Valid {
			stats.NullFraction = nullFrac.Float64
			stats.AvgWidth = int(avgWidth.Int64)
			stats.DistinctValues = int(distinctVals.Float64)
			col.Statistics = &stats
		}

		columns = append(columns, col)
	}

	return columns, nil
}
func (a *Analyzer) getIndexes(schemaName, tableName string) ([]Index, error) {
	query := `
		SELECT 
			i.relname as index_name,
			am.amname as index_type,
			array_to_string(array_agg(a.attname ORDER BY k.nr), ',') as column_names,
			ix.indisunique as is_unique,
			pg_get_expr(ix.indpred, ix.indrelid) as predicate
		FROM pg_class t
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_am am ON i.relam = am.oid
		JOIN pg_attribute a ON a.attrelid = t.oid
		JOIN unnest(ix.indkey) WITH ORDINALITY AS k(attnum, nr)
			ON a.attnum = k.attnum
		WHERE n.nspname = $1
		AND t.relname = $2
		GROUP BY i.relname, am.amname, ix.indisunique, ix.indpred, ix.indrelid
		ORDER BY i.relname`
slog.Info("query", slog.Any("query", query))
	rows, err := a.db.Query(query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []Index
	for rows.Next() {
		var idx Index
		var predicate sql.NullString
		var colNames string

		err := rows.Scan(
			&idx.Name,
			&idx.Type,
			&colNames,
			&idx.Unique,
			&predicate,
		)
		if err != nil {
			return nil, err
		}

		// Converte a string para []string
		if colNames != "" {
			idx.Columns = strings.Split(colNames, ",")
		}

		if predicate.Valid {
			idx.Predicate = predicate.String
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}

func (a *Analyzer) getForeignKeys(schemaName, tableName string) ([]ForeignKey, error) {
	query := `
		SELECT
			c.conname as constraint_name,
			array_agg(col.attname ORDER BY u.attposition) as column_names,
			nf.nspname as foreign_schema,
			tf.relname as foreign_table,
			array_agg(fcol.attname ORDER BY u.attposition) as foreign_columns,
			c.confupdtype,
			c.confdeltype,
			c.condeferrable
		FROM pg_constraint c
		JOIN pg_namespace n ON n.oid = c.connamespace
		JOIN pg_class t ON t.oid = c.conrelid
		JOIN pg_class tf ON tf.oid = c.confrelid
		JOIN pg_namespace nf ON nf.oid = tf.relnamespace
		JOIN pg_attribute col ON col.attrelid = t.oid
		JOIN pg_attribute fcol ON fcol.attrelid = tf.oid
		JOIN (
			SELECT oid, unnest(conkey) as conkey, 
				   unnest(confkey) as confkey,
				   generate_series(1, array_length(conkey, 1)) as attposition
			FROM pg_constraint
		) u ON u.oid = c.oid AND col.attnum = u.conkey AND fcol.attnum = u.confkey
		WHERE n.nspname = $1
		AND t.relname = $2
		AND c.contype = 'f'
		GROUP BY c.conname, nf.nspname, tf.relname, c.confupdtype, c.confdeltype, c.condeferrable
		ORDER BY c.conname`

	rows, err := a.db.Query(query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		var updateType, deleteType string
		var colNames pq.StringArray
		var refColNames pq.StringArray

		err := rows.Scan(
			&fk.Name,
			&colNames,
			&fk.RefSchema,
			&fk.RefTable,
			&refColNames,
			&updateType,
			&deleteType,
			&fk.Deferrable,
		)
		if err != nil {
			return nil, err
		}

		fk.Columns = []string(colNames)
		fk.RefColumns = []string(refColNames)

		// Converte tipos de ação
		fk.OnUpdate = convertActionType(updateType)
		fk.OnDelete = convertActionType(deleteType)

		fks = append(fks, fk)
	}

	return fks, nil
}


func (a *Analyzer) getViews(schemaName string) ([]View, error) {
	query := `
		SELECT 
			v.table_name,
			pg_get_userbyid(c.relowner) as owner,
			pg_get_viewdef(c.oid) as view_definition,
			obj_description(c.oid) as description
		FROM information_schema.views v
		JOIN pg_class c ON c.relname = v.table_name 
			AND c.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = v.table_schema)
		WHERE v.table_schema = $1
		ORDER BY v.table_name`
slog.Info("query", slog.Any("query", query))
	rows, err := a.db.Query(query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []View
	for rows.Next() {
		var view View
		var comment sql.NullString

		view.Schema = schemaName
		err := rows.Scan(
			&view.Name,
			&view.Owner,
			&view.Query,
			&comment,
		)
		if err != nil {
			return nil, err
		}

		if comment.Valid {
			view.Comment = comment.String
		}

		// Obtém colunas da view
		columns, err := a.getColumns(schemaName, view.Name)
		if err != nil {
			return nil, err
		}
		view.Columns = columns

		views = append(views, view)
	}

	return views, nil
}
func (a *Analyzer) getMatViews(schemaName string) ([]MatView, error) {
	query := `
		SELECT 
			c.relname as matview_name,
			pg_get_userbyid(c.relowner) as owner,
			pg_get_viewdef(c.oid) as view_definition,
			obj_description(c.oid) as description
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN (
			SELECT schemaname, matviewname, 
				   pg_size_pretty(pg_relation_size(format('%I.%I', schemaname, matviewname)::regclass)) as size
			FROM pg_matviews
		) s ON s.schemaname = n.nspname AND s.matviewname = c.relname
		WHERE n.nspname = $1
		AND c.relkind = 'm'
		ORDER BY c.relname`

	rows, err := a.db.Query(query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matViews []MatView
	for rows.Next() {
		var mv MatView
		var comment sql.NullString

		mv.Schema = schemaName
		err := rows.Scan(
			&mv.Name,
			&mv.Owner,
			&mv.Query,
			&comment,
		)
		if err != nil {
			return nil, err
		}

		if comment.Valid {
			mv.Comment = comment.String
		}

		// Obtém colunas da view materializada
		columns, err := a.getColumns(schemaName, mv.Name)
		if err != nil {
			return nil, err
		}
		mv.Columns = columns

		// Obtém índices da view materializada
		indexes, err := a.getIndexes(schemaName, mv.Name)
		if err != nil {
			return nil, err
		}
		mv.Indexes = indexes

		matViews = append(matViews, mv)
	}

	return matViews, nil
}

func (a *Analyzer) getFunctions(schemaName string) ([]Function, error) {
	query := `
		SELECT 
			p.proname as function_name,
			pg_get_userbyid(p.proowner) as owner,
			pg_get_function_arguments(p.oid) as arguments,
			pg_get_function_result(p.oid) as result_type,
			l.lanname as language,
			p.prosrc as source,
			obj_description(p.oid) as description,
			p.provolatile,
			p.prosecdef,
			CASE WHEN p.proargmodes IS NULL THEN '' ELSE array_to_string(p.proargmodes, ',') END as arg_modes,
			CASE WHEN p.proargnames IS NULL THEN '' ELSE array_to_string(p.proargnames, ',') END as arg_names,
			CASE WHEN p.proargdefaults IS NULL THEN '' ELSE '' END as arg_defaults
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		JOIN pg_language l ON p.prolang = l.oid
		WHERE n.nspname = COALESCE($1, n.nspname)
		AND p.prokind = 'f'
		ORDER BY p.proname`
slog.Info("query", slog.Any("query", query))
	rows, err := a.db.Query(query, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var functions []Function
	for rows.Next() {
		var fn Function
		var comment, argModes, argNames, argDefaults, arguments sql.NullString
		var volatile string
		var securityDefiner bool

		fn.Schema = schemaName
		err := rows.Scan(
			&fn.Name,
			&fn.Owner,
			&arguments,
			&fn.ReturnType,
			&fn.Language,
			&fn.Source,
			&comment,
			&volatile,
			&securityDefiner,
			&argModes,
			&argNames,
			&argDefaults,
		)
		if err != nil {
			return nil, err
		}

		if comment.Valid {
			fn.Comment = comment.String
		}

		// Converte volatilidade
		fn.Volatile = convertVolatile(volatile)
		fn.Security = convertSecurity(securityDefiner)

		// Parse argumentos
		fn.Arguments = parseArguments(
			arguments.String, 
			argModes.String, 
			argNames.String, 
			argDefaults.String,
		)

		functions = append(functions, fn)
	}

	return functions, nil
}

func convertActionType(actionType string) string {
	switch actionType {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return ""
	}
}

func convertVolatile(volatile string) string {
	switch volatile {
	case "i":
		return "IMMUTABLE"
	case "s":
		return "STABLE"
	case "v":
		return "VOLATILE"
	default:
		return "VOLATILE"
	}
}

func convertSecurity(securityDefiner bool) string {
	if securityDefiner {
		return "DEFINER"
	}
	return "INVOKER"
}

func parseArguments(argStr, modesStr, namesStr, defaultsStr string) []Argument {
	types := strings.Split(argStr, ",")
	modes := strings.Split(modesStr, ",")
	names := strings.Split(namesStr, ",")
	defaults := strings.Split(defaultsStr, ",")

	var args []Argument
	for i := 0; i < len(types); i++ {
		arg := Argument{
			Type: strings.TrimSpace(types[i]),
		}

		if i < len(modes) && modes[i] != "" {
			arg.Mode = convertMode(modes[i])
		} else {
			arg.Mode = "IN"
		}

		if i < len(names) && names[i] != "" {
			arg.Name = names[i]
		}

		if i < len(defaults) && defaults[i] != "" {
			arg.Default = defaults[i]
		}

		args = append(args, arg)
	}

	return args
}

func convertMode(mode string) string {
	switch mode {
	case "i":
		return "IN"
	case "o":
		return "OUT"
	case "b":
		return "INOUT"
	case "v":
		return "VARIADIC"
	case "t":
		return "TABLE"
	default:
		return "IN"
	}
}