// internal/postgres/types.go
package postgres

type Database struct {
    Name      string      `json:"name"`
    Schemas   []Schema    `json:"schemas"`
    Functions []Function  `json:"functions"`
}

type Schema struct {
    Name      string      `json:"name"`
    Tables    []Table     `json:"tables"`
    Views     []View      `json:"views"`
    MatViews  []MatView   `json:"materialized_views"`
    Functions []Function  `json:"functions"`
}

type Table struct {
    Schema      string       `json:"schema"`
    Name        string       `json:"name"`
    Columns     []Column     `json:"columns"`
    Indexes     []Index      `json:"indexes"`
    ForeignKeys []ForeignKey `json:"foreign_keys"`
    Comment     string       `json:"comment,omitempty"`
    Owner       string       `json:"owner"`
    Inherits    []string    `json:"inherits,omitempty"`
}

type Column struct {
    Name        string  `json:"name"`
    Type        string  `json:"type"`
    Nullable    bool    `json:"nullable"`
    PrimaryKey  bool    `json:"primary_key"`
    Default     string  `json:"default,omitempty"`
    Comment     string  `json:"comment,omitempty"`
    Statistics  *ColumnStats `json:"statistics,omitempty"`
}

type ColumnStats struct {
    NullFraction float64 `json:"null_fraction"`
    AvgWidth     int     `json:"avg_width"`
    DistinctValues int   `json:"distinct_values"`
}

type Index struct {
    Name      string   `json:"name"`
    Columns   []string `json:"columns"`
    Type      string   `json:"type"` // btree, hash, gin, gist, etc
    Unique    bool     `json:"unique"`
    Predicate string   `json:"predicate,omitempty"` // WHERE clause
}

type ForeignKey struct {
    Name          string   `json:"name"`
    Columns       []string `json:"columns"`
    RefSchema     string   `json:"ref_schema"`
    RefTable      string   `json:"ref_table"`
    RefColumns    []string `json:"ref_columns"`
    OnDelete      string   `json:"on_delete,omitempty"`
    OnUpdate      string   `json:"on_update,omitempty"`
    Deferrable    bool     `json:"deferrable"`
}

type View struct {
    Schema    string    `json:"schema"`
    Name      string    `json:"name"`
    Columns   []Column  `json:"columns"`
    Query     string    `json:"query"`
    Comment   string    `json:"comment,omitempty"`
    Owner     string    `json:"owner"`
}

type MatView struct {
    Schema      string    `json:"schema"`
    Name        string    `json:"name"`
    Columns     []Column  `json:"columns"`
    Query       string    `json:"query"`
    Comment     string    `json:"comment,omitempty"`
    Owner       string    `json:"owner"`
    Indexes     []Index   `json:"indexes"`
    Refreshed   string    `json:"refreshed_at,omitempty"`
}

type Function struct {
    Schema      string        `json:"schema"`
    Name        string        `json:"name"`
    Arguments   []Argument    `json:"arguments"`
    ReturnType  string        `json:"return_type"`
    Language    string        `json:"language"`
    Source      string        `json:"source"`
    Comment     string        `json:"comment,omitempty"`
    Owner       string        `json:"owner"`
    Volatile    string        `json:"volatile"` // IMMUTABLE, STABLE, VOLATILE
    Security    string        `json:"security"` // INVOKER, DEFINER
}

type Argument struct {
    Name    string `json:"name"`
    Type    string `json:"type"`
    Mode    string `json:"mode"` // IN, OUT, INOUT, VARIADIC
    Default string `json:"default,omitempty"`
}