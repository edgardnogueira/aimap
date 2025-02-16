// internal/nextjs/types.go
package nextjs

type Project struct {
    Name        string       `json:"name"`
    Components  []Component  `json:"components"`
    Pages       []Page       `json:"pages"`
    StateModules []StateModule `json:"state_modules"`
    Layouts     []Layout     `json:"layouts"`
    APIs        []API        `json:"apis"`
}

type Component struct {
    Name       string       `json:"name"`
    Path       string       `json:"path"`
    Props      []Prop      `json:"props"`
    Hooks      []Hook      `json:"hooks"`
    Children   []Component `json:"children,omitempty"`
    IsServer   bool        `json:"is_server"`
    IsClient   bool        `json:"is_client"`
    Type       string      `json:"type"` // functional, class, etc
}

type Prop struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Required bool   `json:"required"`
    Default  string `json:"default,omitempty"`
}

type Hook struct {
    Name       string   `json:"name"`
    Type       string   `json:"type"` // useState, useEffect, useCallback, etc
    Dependencies []string `json:"dependencies,omitempty"`
}

type Page struct {
    Route      string     `json:"route"`
    Path       string     `json:"path"`
    Layout     string     `json:"layout,omitempty"`
    Components []string   `json:"components"`
    Params     []Param    `json:"params,omitempty"`
    APIs       []string   `json:"apis,omitempty"`
    IsStatic   bool       `json:"is_static"`
    IsDynamic  bool       `json:"is_dynamic"`
}

type Param struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Required bool   `json:"required"`
}

type Layout struct {
    Name       string     `json:"name"`
    Path       string     `json:"path"`
    Components []string   `json:"components"`
    IsRoot     bool       `json:"is_root"`
}

type StateModule struct {
    Name       string     `json:"name"`
    Path       string     `json:"path"`
    Type       string     `json:"type"` // redux, zustand, jotai, etc
    Actions    []Action   `json:"actions,omitempty"`
    Slices     []Slice    `json:"slices,omitempty"`
    Atoms      []Atom     `json:"atoms,omitempty"`
}

type Action struct {
    Name       string   `json:"name"`
    Type       string   `json:"type"`
    Payload    string   `json:"payload,omitempty"`
}

type Slice struct {
    Name       string    `json:"name"`
    State      string    `json:"state"`
    Reducers   []string  `json:"reducers"`
}

type Atom struct {
    Name       string    `json:"name"`
    Type       string    `json:"type"`
    Default    string    `json:"default,omitempty"`
}

type API struct {
    Route      string    `json:"route"`
    Path       string    `json:"path"`
    Method     string    `json:"method"`
    Handler    string    `json:"handler"`
    Middleware []string  `json:"middleware,omitempty"`
}