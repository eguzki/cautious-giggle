package handlers

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
	"github.com/eguzki/cautious-giggle/pkg/utils"
)

type Operation struct {
	Method      string
	Path        string
	Security    string
	OperationID string
	Target      string
}

type NewAPIData struct {
	ServiceName string
	Operations  []Operation
}

type NewApiHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &NewApiHandler{}

func (a *NewApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["service"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "url param service not found", http.StatusBadRequest)
		return
	}

	serviceName := keys[0]
	serviceKey := client.ObjectKey{Name: serviceName, Namespace: "default"}
	serviceObj := &corev1.Service{}
	err := a.K8sClient.Get(context.Background(), serviceKey, serviceObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var newAPIData *NewAPIData
	var templateName string
	oasConfigmapName, ok := serviceObj.Annotations[utils.KuadrantDiscoveryAnnotationOASConfigMap]
	if ok {
		var err error
		newAPIData, err = a.readOASFromConfigMap(oasConfigmapName, serviceObj.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templateName = "newapi.html.tmpl"
	} else {
		newAPIData = &NewAPIData{
			ServiceName: serviceObj.Name,
		}
		templateName = "newapi-oasform.html.tmpl"
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, newAPIData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *NewApiHandler) readOASFromConfigMap(oasConfigmapName, serviceName string) (*NewAPIData, error) {
	oasConfigmap := &corev1.ConfigMap{}
	oasConfigMapKey := client.ObjectKey{Name: oasConfigmapName, Namespace: "default"}
	err := a.K8sClient.Get(context.Background(), oasConfigMapKey, oasConfigmap)
	if err != nil {
		return nil, err
	}

	oasContent, ok := oasConfigmap.Data["openapi.yaml"]
	if !ok {
		return nil, errors.New("oas configmap is missing the openapi.yaml entry")
	}

	openapiLoader := openapi3.NewLoader()
	doc, err := openapiLoader.LoadFromData([]byte(oasContent))
	if err != nil {
		return nil, err
	}

	err = doc.Validate(openapiLoader.Context)
	if err != nil {
		return nil, err
	}

	newAPIData := &NewAPIData{
		ServiceName: serviceName,
	}

	for path, pathItem := range doc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			secReqsP := utils.OpenAPIOperationSecRequirements(doc, operation)

			var security string = "None"

			if secReqsP != nil {
				for _, secReq := range *secReqsP {
					for secSchemeName := range secReq {
						secSchemeI, err := doc.Components.SecuritySchemes.JSONLookup(secSchemeName)
						if err != nil {
							return nil, err
						}

						secScheme := secSchemeI.(*openapi3.SecurityScheme) // panic if assertion fails
						security = secScheme.Type
					}
				}
			}

			newAPIData.Operations = append(newAPIData.Operations, Operation{
				Method:   opVerb,
				Path:     path,
				Security: security,
			})
		}
	}

	return newAPIData, nil
}
