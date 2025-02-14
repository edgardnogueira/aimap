// pkg/kubedoc/parser.go
package kubedoc

import (
	"fmt"
	"io/ioutil"
	"log/slog"
	"path/filepath"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)
func init() {
    // Registra os tipos que queremos ser capazes de deserializar
    appsv1.AddToScheme(scheme.Scheme)
    corev1.AddToScheme(scheme.Scheme)
    networkingv1.AddToScheme(scheme.Scheme)
}


func (p *Parser) ParseDirectory(dirPath string) error {
    files, err := ioutil.ReadDir(dirPath)
    if err != nil {
        return fmt.Errorf("erro ao ler diretório: %v", err)
    }

    for _, file := range files {
        if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
            slog.Info("Processando arquivo", "filename", file.Name())
            if err := p.ParseFile(filepath.Join(dirPath, file.Name())); err != nil {
                slog.Error("Erro ao processar arquivo", "filename", file.Name(), "error", err)
                continue
            }
        }
    }

    return nil
}

func (p *Parser) ParseFile(filename string) error {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("erro ao ler arquivo: %v", err)
    }

    // Divide em documentos YAML múltiplos
    documents := strings.Split(string(data), "---")
    
    decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
    
    for i, doc := range documents {
        if strings.TrimSpace(doc) == "" {
            continue
        }

        obj, kind, err := decode([]byte(doc), nil, nil)
        if err != nil {
            slog.Error("Erro ao decodificar documento", 
                "filename", filename,
                "doc_index", i,
                "error", err)
            continue
        }

        if err := p.parseResource(obj, kind.Kind); err != nil {
            slog.Error("Erro ao processar recurso", 
                "kind", kind.Kind,
                "error", err)
            continue
        }
    }

    return nil
}

func (p *Parser) parseResource(obj runtime.Object, kind string) error {
    // Extrai metadados comuns
    metadata, err := meta.Accessor(obj)
    if err != nil {
        return err
    }

    node := &ResourceNode{
        Name:      metadata.GetName(),
        Kind:      kind,
        Namespace: metadata.GetNamespace(),
        Labels:    metadata.GetLabels(),
    }

    // Converte o spec para map[string]interface{}
    rawObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
    if err == nil {
        if spec, ok := rawObj["spec"].(map[string]interface{}); ok {
            node.RawSpec = spec
        }
    }

    // Adiciona ao mapa de recursos
    p.resources[node.Name] = node

    // Analisa relações
    p.parseRelations(node, obj)

    return nil
}

func (p *Parser) parseRelations(node *ResourceNode, obj runtime.Object) {
    switch obj := obj.(type) {
    case *appsv1.Deployment:
        p.parseDeploymentRelations(node, obj)
    case *v1.Service:
        p.parseServiceRelations(node, obj)
    case *networkingv1.Ingress:
        p.parseIngressRelations(node, obj)
    }
}
// Adicione estes métodos à struct Parser

func (p *Parser) parseDeploymentRelations(node *ResourceNode, deployment *appsv1.Deployment) {
    if deployment.Spec.Template.Labels != nil {
        node.Labels = deployment.Spec.Template.Labels
    }

    // Relaciona com serviços que podem selecionar este deployment
    for _, otherResource := range p.resources {
        if otherResource.Kind == "Service" {
            if svc, ok := otherResource.RawSpec["selector"].(map[string]interface{}); ok {
                matches := true
                for k, v := range svc {
                    if labelV, exists := node.Labels[k]; !exists || labelV != v {
                        matches = false
                        break
                    }
                }
                if matches {
                    relation := Relation{
                        FromName: otherResource.Name,
                        ToName:   node.Name,
                        Kind:     "selects",
                    }
                    otherResource.Relations = append(otherResource.Relations, relation)
                }
            }
        }
    }
}

func (p *Parser) parseServiceRelations(node *ResourceNode, service *corev1.Service) {
    if service.Spec.Selector != nil {
        // Procura deployments/statefulsets que são selecionados por este serviço
        for _, otherResource := range p.resources {
            if otherResource.Kind == "Deployment" || otherResource.Kind == "StatefulSet" {
                // Removemos a asserção de tipo desnecessária
                matches := true
                for k, v := range service.Spec.Selector {
                    if labelV, exists := otherResource.Labels[k]; !exists || labelV != v {
                        matches = false
                        break
                    }
                }
                if matches {
                    relation := Relation{
                        FromName: node.Name,
                        ToName:   otherResource.Name,
                        Kind:     "selects",
                    }
                    node.Relations = append(node.Relations, relation)
                }
            }
        }
    }
}

func (p *Parser) parseIngressRelations(node *ResourceNode, ingress *networkingv1.Ingress) {
    for _, rule := range ingress.Spec.Rules {
        if rule.HTTP != nil {
            for _, path := range rule.HTTP.Paths {
                if path.Backend.Service != nil {
                    relation := Relation{
                        FromName: node.Name,
                        ToName:   path.Backend.Service.Name,
                        Kind:     "routes",
                    }
                    node.Relations = append(node.Relations, relation)
                }
            }
        }
    }
}