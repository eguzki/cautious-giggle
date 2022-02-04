package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
	"github.com/eguzki/cautious-giggle/pkg/utils"
)

type NewRLPlanOperation struct {
	Operation   string
	OperationID string
}

type NewRLPlanData struct {
	APIName          string
	APIDomain        string
	AuthOperations   []NewRLPlanOperation
	UnAuthOperations []NewRLPlanOperation
}

type NewRateLimitPlanHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &NewRateLimitPlanHandler{}

func (a *NewRateLimitPlanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["api"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "url param api not found", http.StatusBadRequest)
		return
	}

	apiName := keys[0]
	apiObj := &gigglev1alpha1.Api{}
	apiKey := client.ObjectKey{Name: apiName, Namespace: "default"}
	err := a.K8sClient.Get(context.Background(), apiKey, apiObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &NewRLPlanData{
		APIName:   apiName,
		APIDomain: apiObj.Spec.PublicDomain,
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
			data.UnAuthOperations = append(data.UnAuthOperations, NewRLPlanOperation{
				Operation:   fmt.Sprintf("%s %s", opVerb, path),
				OperationID: operation.OperationID,
			})

			if operationIsSecured(doc, operation) {
				data.AuthOperations = append(data.AuthOperations, NewRLPlanOperation{
					Operation:   fmt.Sprintf("%s %s", opVerb, path),
					OperationID: operation.OperationID,
				})
			}
		}
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "newplan.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func operationIsSecured(doc *openapi3.T, operation *openapi3.Operation) bool {
	secReqsP := utils.OpenAPIOperationSecRequirements(doc, operation)

	if secReqsP == nil {
		return false
	}
	if len(*secReqsP) == 0 {
		return false
	}
	for _, secReq := range *secReqsP {
		if len(secReq) != 0 {
			return true
		}
	}

	return false
}
