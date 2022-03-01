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

type APIHandler struct {
	K8sClient client.Client
}

type Plan struct {
	ID string
}

type APIData struct {
	Name                  string
	ServiceName           string
	Description           string
	PublicDomain          string
	PathMatchType         string
	Operations            []Operation
	Plans                 []Plan
	RateLimitOperations   []*PlanOperation
	UnAuthGlobalDaily     string
	UnAuthGlobalMonthly   string
	UnAuthGlobalYearly    string
	UnAuthRemoteIPDaily   string
	UnAuthRemoteIPMonthly string
	UnAuthRemoteIPYearly  string
}

var _ http.Handler = &APIHandler{}

func (a *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["id"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "url param id not found", http.StatusBadRequest)
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
		Name:                  apiName,
		ServiceName:           apiObj.Spec.ServiceName,
		Description:           apiObj.Spec.Description,
		PublicDomain:          apiObj.Spec.PublicDomain,
		PathMatchType:         apiObj.Spec.PathMatchType,
		UnAuthGlobalDaily:     "0",
		UnAuthGlobalMonthly:   "0",
		UnAuthGlobalYearly:    "0",
		UnAuthRemoteIPDaily:   "0",
		UnAuthRemoteIPMonthly: "0",
		UnAuthRemoteIPYearly:  "0",
	}

	for planName := range apiObj.Spec.Plans {
		aPIData.Plans = append(aPIData.Plans, Plan{ID: planName})
	}

	if apiObj.Spec.GetUnAuthRateLimit().GetGlobal().Daily != nil {
		aPIData.UnAuthGlobalDaily = fmt.Sprint(*apiObj.Spec.GetUnAuthRateLimit().GetGlobal().Daily)
	}
	if apiObj.Spec.GetUnAuthRateLimit().GetGlobal().Monthly != nil {
		aPIData.UnAuthGlobalMonthly = fmt.Sprint(*apiObj.Spec.GetUnAuthRateLimit().GetGlobal().Monthly)
	}
	if apiObj.Spec.GetUnAuthRateLimit().GetGlobal().Yearly != nil {
		aPIData.UnAuthGlobalYearly = fmt.Sprint(*apiObj.Spec.GetUnAuthRateLimit().GetGlobal().Yearly)
	}

	if apiObj.Spec.GetUnAuthRateLimit().GetRemoteIP().Daily != nil {
		aPIData.UnAuthRemoteIPDaily = fmt.Sprint(*apiObj.Spec.GetUnAuthRateLimit().GetRemoteIP().Daily)
	}
	if apiObj.Spec.GetUnAuthRateLimit().GetRemoteIP().Monthly != nil {
		aPIData.UnAuthRemoteIPMonthly = fmt.Sprint(*apiObj.Spec.GetUnAuthRateLimit().GetRemoteIP().Monthly)
	}
	if apiObj.Spec.GetUnAuthRateLimit().GetRemoteIP().Yearly != nil {
		aPIData.UnAuthRemoteIPYearly = fmt.Sprint(*apiObj.Spec.GetUnAuthRateLimit().GetRemoteIP().Yearly)
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
				Method:      opVerb,
				Path:        path,
				Security:    security,
				OperationID: operation.OperationID,
				Target:      fmt.Sprintf("%s %s", opVerb, path),
			})
			aPIData.RateLimitOperations = append(aPIData.RateLimitOperations, &PlanOperation{
				Operation:   fmt.Sprintf("%s %s", opVerb, path),
				OperationID: operation.OperationID,
				Daily:       "0",
				Monthly:     "0",
				Yearly:      "0",
			})
		}
	}

	for operationID, rlConf := range apiObj.Spec.GetUnAuthRateLimit().Operations {
		for idx, po := range aPIData.RateLimitOperations {
			if aPIData.RateLimitOperations[idx].OperationID == operationID {
				if rlConf != nil && rlConf.Daily != nil {
					po.Daily = fmt.Sprint(*rlConf.Daily)
				}

				if rlConf != nil && rlConf.Monthly != nil {
					po.Monthly = fmt.Sprint(*rlConf.Monthly)
				}

				if rlConf != nil && rlConf.Yearly != nil {
					po.Yearly = fmt.Sprint(*rlConf.Yearly)
				}
			}
		}
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
