// internal/laravel/types.go
package laravel

type Project struct {
    Name       string      `json:"name"`
    Models     []Model     `json:"models"`
    Controllers []Controller `json:"controllers"`
    Routes     []Route     `json:"routes"`
    Migrations []Migration `json:"migrations"`
    Middleware []Middleware `json:"middleware"`
    Providers  []Provider  `json:"providers"`
}

type Model struct {
    Name         string       `json:"name"`
    Table        string       `json:"table"`
    Fillable     []string     `json:"fillable"`
    Hidden       []string     `json:"hidden"`
    Casts        []Cast       `json:"casts"`
    Relationships []Relationship `json:"relationships"`
    Path         string       `json:"path"`
}

type Cast struct {
    Field    string `json:"field"`
    CastType string `json:"cast_type"`
}

type Relationship struct {
    Type      string `json:"type"` // hasOne, hasMany, belongsTo, belongsToMany, etc
    Method    string `json:"method"`
    RelatedModel string `json:"related_model"`
    ForeignKey string `json:"foreign_key,omitempty"`
    LocalKey   string `json:"local_key,omitempty"`
    PivotTable string `json:"pivot_table,omitempty"`
}

type Controller struct {
    Name      string    `json:"name"`
    Path      string    `json:"path"`
    Methods   []Method  `json:"methods"`
    Middleware []string `json:"middleware"`
}

type Method struct {
    Name       string   `json:"name"`
    Parameters []string `json:"parameters"`
    ReturnType string   `json:"return_type"`
    Route      *Route   `json:"route,omitempty"`
}

type Route struct {
    Method     string `json:"method"` // GET, POST, etc
    URI        string `json:"uri"`
    Action     string `json:"action"` // Controller@method
    Name       string `json:"name,omitempty"`
    Middleware []string `json:"middleware"`
}

type Migration struct {
    Name      string    `json:"name"`
    Timestamp string    `json:"timestamp"`
    Table     string    `json:"table"`
    Columns   []Column  `json:"columns"`
    Indexes   []Index   `json:"indexes"`
    Path      string    `json:"path"`
}

type Column struct {
    Name       string `json:"name"`
    Type       string `json:"type"`
    Nullable   bool   `json:"nullable"`
    Default    string `json:"default,omitempty"`
    Unique     bool   `json:"unique"`
    Primary    bool   `json:"primary"`
    References *Reference `json:"references,omitempty"`
}

type Reference struct {
    Table    string `json:"table"`
    Column   string `json:"column"`
    OnDelete string `json:"on_delete,omitempty"`
    OnUpdate string `json:"on_update,omitempty"`
}

type Index struct {
    Name    string   `json:"name"`
    Columns []string `json:"columns"`
    Type    string   `json:"type"` // index, unique, primary
}

type Middleware struct {
    Name      string `json:"name"`
    Path      string `json:"path"`
    Global    bool   `json:"global"`
    Group     string `json:"group,omitempty"`
}

type Provider struct {
    Name      string `json:"name"`
    Path      string `json:"path"`
    Type      string `json:"type"` // Service Provider type
    Deferred  bool   `json:"deferred"`
}