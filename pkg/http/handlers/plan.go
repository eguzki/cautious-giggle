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
)

type PlanOperation struct {
	Operation   string
	OperationID string
	Daily       string
	Monthly     string
	Eternity    string
}

type PlanData struct {
	APIName                string
	APIDomain              string
	Name                   string
	Description            string
	UnAuthGlobalDaily      string
	UnAuthGlobalMonthly    string
	UnAuthGlobalEternity   string
	UnAuthRemoteIPDaily    string
	UnAuthRemoteIPMonthly  string
	UnAuthRemoteIPEternity string
	UnAuthOperations       []*PlanOperation
	AuthGlobalDaily        string
	AuthGlobalMonthly      string
	AuthGlobalEternity     string
	AuthOperations         []*PlanOperation
}

type PlanHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &PlanHandler{}

func (a *PlanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["id"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "url param id not found", http.StatusBadRequest)
		return
	}

	planName := keys[0]

	keys, ok = r.URL.Query()["api"]
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

	plan, ok := apiObj.Spec.Plans[planName]
	if !ok || plan == nil {
		http.Error(w, "url param id not found in the api object plans", http.StatusBadRequest)
		return
	}

	data := &PlanData{
		APIName:                apiName,
		APIDomain:              apiObj.Spec.PublicDomain,
		Name:                   planName,
		Description:            plan.Description,
		UnAuthGlobalDaily:      "---",
		UnAuthGlobalMonthly:    "---",
		UnAuthGlobalEternity:   "---",
		UnAuthRemoteIPDaily:    "---",
		UnAuthRemoteIPMonthly:  "---",
		UnAuthRemoteIPEternity: "---",
		AuthGlobalDaily:        "---",
		AuthGlobalMonthly:      "---",
		AuthGlobalEternity:     "---",
	}

	// Initialize
	for path, pathItem := range doc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			data.UnAuthOperations = append(data.UnAuthOperations, &PlanOperation{
				Operation:   fmt.Sprintf("%s %s", opVerb, path),
				OperationID: operation.OperationID,
				Daily:       "---",
				Monthly:     "---",
				Eternity:    "---",
			})

			if operationIsSecured(doc, operation) {
				data.AuthOperations = append(data.AuthOperations, &PlanOperation{
					Operation:   fmt.Sprintf("%s %s", opVerb, path),
					OperationID: operation.OperationID,
					Daily:       "---",
					Monthly:     "---",
					Eternity:    "---",
				})
			}
		}
	}

	if plan.GetUnAuth().GetGlobal().Daily != nil {
		data.UnAuthGlobalDaily = fmt.Sprint(*plan.GetUnAuth().GetGlobal().Daily)
	}
	if plan.GetUnAuth().GetGlobal().Monthly != nil {
		data.UnAuthGlobalMonthly = fmt.Sprint(*plan.GetUnAuth().GetGlobal().Monthly)
	}
	if plan.GetUnAuth().GetGlobal().Eternity != nil {
		data.UnAuthGlobalEternity = fmt.Sprint(*plan.GetUnAuth().GetGlobal().Eternity)
	}

	if plan.GetUnAuth().GetRemoteIP().Daily != nil {
		data.UnAuthRemoteIPDaily = fmt.Sprint(*plan.GetUnAuth().GetRemoteIP().Daily)
	}
	if plan.GetUnAuth().GetRemoteIP().Monthly != nil {
		data.UnAuthRemoteIPMonthly = fmt.Sprint(*plan.GetUnAuth().GetRemoteIP().Monthly)
	}
	if plan.GetUnAuth().GetRemoteIP().Eternity != nil {
		data.UnAuthRemoteIPEternity = fmt.Sprint(*plan.GetUnAuth().GetRemoteIP().Eternity)
	}

	if plan.GetAuth().GetGlobal().Daily != nil {
		data.AuthGlobalDaily = fmt.Sprint(*plan.GetAuth().GetGlobal().Daily)
	}
	if plan.GetAuth().GetGlobal().Monthly != nil {
		data.AuthGlobalMonthly = fmt.Sprint(*plan.GetAuth().GetGlobal().Monthly)
	}
	if plan.GetAuth().GetGlobal().Eternity != nil {
		data.AuthGlobalEternity = fmt.Sprint(*plan.GetAuth().GetGlobal().Eternity)
	}

	for operationID, operationPlan := range plan.GetUnAuth().Operations {
		for idx, po := range data.UnAuthOperations {
			if data.UnAuthOperations[idx].OperationID == operationID {
				if operationPlan != nil && operationPlan.Daily != nil {
					po.Daily = fmt.Sprint(*operationPlan.Daily)
				}

				if operationPlan != nil && operationPlan.Monthly != nil {
					po.Monthly = fmt.Sprint(*operationPlan.Monthly)
				}

				if operationPlan != nil && operationPlan.Eternity != nil {
					po.Eternity = fmt.Sprint(*operationPlan.Eternity)
				}
			}
		}
	}

	for operationID, operationPlan := range plan.GetAuth().Operations {
		for idx, po := range data.AuthOperations {
			if data.AuthOperations[idx].OperationID == operationID {
				if operationPlan != nil && operationPlan.Daily != nil {
					po.Daily = fmt.Sprint(*operationPlan.Daily)
				}

				if operationPlan != nil && operationPlan.Monthly != nil {
					po.Monthly = fmt.Sprint(*operationPlan.Monthly)
				}

				if operationPlan != nil && operationPlan.Eternity != nil {
					po.Eternity = fmt.Sprint(*operationPlan.Eternity)
				}
			}
		}
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "plan.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
