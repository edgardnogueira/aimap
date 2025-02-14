package kubedoc

// Resources representa a coleção de recursos Kubernetes
type Resources struct {
    Resources []Resource `json:"resources" yaml:"resources"`
}

// Resource representa um recurso Kubernetes individual
type Resource struct {
    Name      string            `json:"name" yaml:"name"`
    Kind      string            `json:"kind" yaml:"kind"`
    Namespace string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
    Labels    map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
    Relations []Relation        `json:"relations,omitempty" yaml:"relations,omitempty"`
}

// Relation representa uma relação entre recursos Kubernetes
type Relation struct {
    FromName string `json:"fromName" yaml:"fromName"`
    ToName   string `json:"toName" yaml:"toName"`
    Kind     string `json:"kind" yaml:"kind"` // e.g., "selects", "mounts", "references"
}


// Parser is responsible for parsing Kubernetes resource files
type Parser struct {
    resources map[string]*ResourceNode
}

// NewParser creates a new Kubernetes resource parser
func NewParser() *Parser {
    return &Parser{
        resources: make(map[string]*ResourceNode),
    }
}

// ResourceNode represents a Kubernetes resource in the dependency graph
type ResourceNode struct {
    Name      string
    Kind      string
    Namespace string
    Labels    map[string]string
    Relations []Relation
    RawSpec   map[string]interface{}
}

// GetResources returns all parsed resources
func (p *Parser) GetResources() []*ResourceNode {
    var resources []*ResourceNode
    for _, resource := range p.resources {
        resources = append(resources, resource)
    }
    return resources
}

// GetResource returns a specific resource by name
func (p *Parser) GetResource(name string) *ResourceNode {
    return p.resources[name]
}

// AddResource adds a new resource to the parser
func (p *Parser) AddResource(resource *ResourceNode) {
    p.resources[resource.Name] = resource
}

// RemoveResource removes a resource from the parser
func (p *Parser) RemoveResource(name string) {
    delete(p.resources, name)
}