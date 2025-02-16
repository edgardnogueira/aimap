// internal/mysql/types.go
package mysql

type Database struct {
	Name    string   `json:"name"`
	Tables  []Table  `json:"tables"`
	Views   []View   `json:"views"`
}

type Table struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	Indexes     []Index      `json:"indexes"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
	Comment     string       `json:"comment,omitempty"`
}

type View struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
	Query   string   `json:"query"`
	Comment string   `json:"comment,omitempty"`
}

type Column struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	PrimaryKey bool   `json:"primary_key"`
	Default    string `json:"default,omitempty"`
	Extra      string `json:"extra,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"` // BTREE, HASH, etc
	Unique  bool     `json:"unique"`
}

type ForeignKey struct {
	Name            string   `json:"name"`
	Columns         []string `json:"columns"`
	RefTable        string   `json:"ref_table"`
	RefColumns      []string `json:"ref_columns"`
	OnDelete        string   `json:"on_delete,omitempty"`
	OnUpdate        string   `json:"on_update,omitempty"`
}