package handlers

import (
	"context"
	"html/template"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
	"github.com/eguzki/cautious-giggle/pkg/utils"
)

type APIHandler struct {
	K8sClient client.Client
}

type Plan struct {
	ID string
}

type APIData struct {
	Name          string
	ServiceName   string
	Description   string
	PublicDomain  string
	PathMatchType string
	Gateway       string
	Operations    []Operation
	Plans         []Plan
	Gateways      []Gateway
}

var _ http.Handler = &APIHandler{}

func (a *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["id"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "url param api not found", http.StatusBadRequest)
		return
	}

	apiName := keys[0]
	apikey := client.ObjectKey{Name: apiName, Namespace: "default"}
	apiObj := &gigglev1alpha1.Api{}
	err := a.K8sClient.Get(context.Background(), apikey, apiObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	aPIData := APIData{
		Name:          apiName,
		ServiceName:   apiObj.Spec.ServiceName,
		Description:   apiObj.Spec.Description,
		PublicDomain:  apiObj.Spec.PublicDomain,
		PathMatchType: apiObj.Spec.PathMatchType,
	}

	if apiObj.Spec.Gateway != nil {
		aPIData.Gateway = *apiObj.Spec.Gateway
	}

	for planName := range apiObj.Spec.Plans {
		aPIData.Plans = append(aPIData.Plans, Plan{ID: planName})
	}

	openapiLoader := openapi3.NewLoader()
	doc, err := openapiLoader.LoadFromData([]byte(apiObj.Spec.OAS))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = doc.Validate(openapiLoader.Context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}

						secScheme := secSchemeI.(*openapi3.SecurityScheme) // panic if assertion fails
						security = secScheme.Type
					}
				}
			}

			aPIData.Operations = append(aPIData.Operations, Operation{
				Method:   opVerb,
				Path:     path,
				Security: security,
			})
		}
	}

	serviceList := &corev1.ServiceList{}
	err = a.K8sClient.List(context.Background(), serviceList, client.HasLabels{utils.KuadrantGatewayLabel})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for idx := range serviceList.Items {
		aPIData.Gateways = append(aPIData.Gateways, Gateway{Name: serviceList.Items[idx].Name})
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "api.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, aPIData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
