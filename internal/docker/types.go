// internal/docker/types.go
package docker

import (
	"fmt"
	"strings"
)

type Project struct {
    Name        string       `json:"name"`
    Dockerfiles []Dockerfile `json:"dockerfiles"`
    Compose     *Compose     `json:"compose,omitempty"`
}

type Dockerfile struct {
    Path     string    `json:"path"`
    Base     string    `json:"base"`
    Stages   []Stage   `json:"stages,omitempty"`
    Steps    []Step    `json:"steps"`
    Expose   []string  `json:"expose,omitempty"`
    Volumes  []string  `json:"volumes,omitempty"`
    Env      []EnvVar  `json:"env,omitempty"`
    Commands []Command `json:"commands,omitempty"`
}

type Stage struct {
    Name  string  `json:"name"`
    Base  string  `json:"base"`
    Steps []Step  `json:"steps"`
}

type Step struct {
    Type    string   `json:"type"` // RUN, COPY, ADD, etc
    Command string   `json:"command"`
    Args    []string `json:"args,omitempty"`
}

type EnvVar struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

type Command struct {
    Type    string   `json:"type"` // CMD, ENTRYPOINT
    Command string   `json:"command"`
    Args    []string `json:"args,omitempty"`
}

type Compose struct {
    Version  string                  `json:"version" yaml:"version"`
    Services map[string]Service      `json:"services" yaml:"services"`
    Networks map[string]Network      `json:"networks,omitempty" yaml:"networks,omitempty"`
    Volumes  map[string]ComposeVolume `json:"volumes,omitempty" yaml:"volumes,omitempty"`
    Configs  map[string]ComposeConfig `json:"configs,omitempty" yaml:"configs,omitempty"`
    Secrets  map[string]ComposeSecret `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

type Service struct {
    Name          string               `json:"name" yaml:"-"`
    Image         string               `json:"image,omitempty" yaml:"image,omitempty"`
    Build         *BuildConfig         `json:"build,omitempty" yaml:"build,omitempty"`
    Ports         []interface{}        `json:"ports,omitempty" yaml:"ports,omitempty"`    // Pode ser string ou PortMapping
    Environment   []interface{}        `json:"environment,omitempty" yaml:"environment,omitempty"` // Pode ser string ou EnvVar
    Volumes       []interface{}        `json:"volumes,omitempty" yaml:"volumes,omitempty"` // Pode ser string ou VolumeMapping
    Networks      []string             `json:"networks,omitempty" yaml:"networks,omitempty"`
    DependsOn     []string             `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
    Restart       string               `json:"restart,omitempty" yaml:"restart,omitempty"`
    Deploy        *DeployConfig        `json:"deploy,omitempty" yaml:"deploy,omitempty"`
    Healthcheck   *HealthcheckConfig   `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty"`
}

// Função auxiliar para converter interface em VolumeMapping
func parseVolume(v interface{}) VolumeMapping {
    switch val := v.(type) {
    case string:
        parts := strings.Split(val, ":")
        mapping := VolumeMapping{
            Source: parts[0],
            Target: parts[len(parts)-1],
        }
        if len(parts) > 2 {
            mapping.ReadOnly = strings.Contains(parts[1], "ro")
        }
        return mapping
    case map[string]interface{}:
        mapping := VolumeMapping{}
        if src, ok := val["source"]; ok {
            mapping.Source = src.(string)
        }
        if target, ok := val["target"]; ok {
            mapping.Target = target.(string)
        }
        if typ, ok := val["type"]; ok {
            mapping.Type = typ.(string)
        }
        if ro, ok := val["read_only"]; ok {
            mapping.ReadOnly = ro.(bool)
        }
        return mapping
    default:
        return VolumeMapping{}
    }
}

// Função auxiliar para converter interface em PortMapping
func parsePort(p interface{}) PortMapping {
    switch val := p.(type) {
    case string:
        parts := strings.Split(val, ":")
        if len(parts) == 1 {
            return PortMapping{Container: parts[0]}
        }
        mapping := PortMapping{
            Host:      parts[0],
            Container: parts[1],
        }
        if len(parts) > 2 {
            mapping.Protocol = parts[2]
        }
        return mapping
    case map[string]interface{}:
        mapping := PortMapping{}
        if target, ok := val["target"]; ok {
            mapping.Container = target.(string)
        }
        if published, ok := val["published"]; ok {
            mapping.Host = published.(string)
        }
        if protocol, ok := val["protocol"]; ok {
            mapping.Protocol = protocol.(string)
        }
        return mapping
    default:
        return PortMapping{}
    }
}

// Função auxiliar para converter interface em EnvVar
func parseEnv(e interface{}) EnvVar {
    switch val := e.(type) {
    case string:
        parts := strings.SplitN(val, "=", 2)
        if len(parts) == 1 {
            return EnvVar{Key: parts[0]}
        }
        return EnvVar{
            Key:   parts[0],
            Value: parts[1],
        }
    case map[string]interface{}:
        env := EnvVar{}
        for k, v := range val {
            env.Key = k
            env.Value = fmt.Sprintf("%v", v)
            break
        }
        return env
    default:
        return EnvVar{}
    }
}

type BuildConfig struct {
    Context    string            `json:"context"`
    Dockerfile string            `json:"dockerfile,omitempty"`
    Args       map[string]string `json:"args,omitempty"`
    Target     string            `json:"target,omitempty"`
}

type PortMapping struct {
    Host      string `json:"host,omitempty"`
    Container string `json:"container"`
    Protocol  string `json:"protocol,omitempty"`
}

type VolumeMapping struct {
    Source      string `json:"source"`
    Target      string `json:"target"`
    Type        string `json:"type,omitempty"` // bind, volume, tmpfs
    ReadOnly    bool   `json:"read_only,omitempty"`
}

type Network struct {
    Name       string   `json:"name"`
    Driver     string   `json:"driver,omitempty"`
    External   bool     `json:"external,omitempty"`
    Attachable bool     `json:"attachable,omitempty"`
    IPAM       *IPAMConfig `json:"ipam,omitempty"`
}

type IPAMConfig struct {
    Driver string     `json:"driver,omitempty"`
    Config []IPAMPool `json:"config,omitempty"`
}

type IPAMPool struct {
    Subnet  string `json:"subnet,omitempty"`
    Gateway string `json:"gateway,omitempty"`
}

type ComposeVolume struct {
    Name     string `json:"name"`
    Driver   string `json:"driver,omitempty"`
    External bool   `json:"external,omitempty"`
}

type ComposeConfig struct {
    Name     string `json:"name"`
    File     string `json:"file,omitempty"`
    External bool   `json:"external,omitempty"`
}

type ComposeSecret struct {
    Name     string `json:"name"`
    File     string `json:"file,omitempty"`
    External bool   `json:"external,omitempty"`
}

type DeployConfig struct {
    Replicas      int             `json:"replicas,omitempty"`
    Resources     ResourceLimits  `json:"resources,omitempty"`
    RestartPolicy string          `json:"restart_policy,omitempty"`
    Placement     []string        `json:"placement,omitempty"`
}

type ResourceLimits struct {
    Limits   *Resources `json:"limits,omitempty"`
    Reservations *Resources `json:"reservations,omitempty"`
}

type Resources struct {
    CPUs    string `json:"cpus,omitempty"`
    Memory  string `json:"memory,omitempty"`
}

type HealthcheckConfig struct {
    Test        []string `json:"test"`
    Interval    string   `json:"interval,omitempty"`
    Timeout     string   `json:"timeout,omitempty"`
    Retries     int      `json:"retries,omitempty"`
    StartPeriod string   `json:"start_period,omitempty"`
}