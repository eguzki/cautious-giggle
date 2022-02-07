package handlers

import (
	"context"
	"fmt"
	"net/http"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UpdateUnAuthRateLimitHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &UpdateUnAuthRateLimitHandler{}

func (a *UpdateUnAuthRateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	apiName := r.FormValue("api")
	if apiName == "" {
		http.Error(w, "form param api not found.", http.StatusBadRequest)
		return
	}
	apiObj := &gigglev1alpha1.Api{}
	apiKey := client.ObjectKey{Name: apiName, Namespace: "default"}
	err = a.K8sClient.Get(context.Background(), apiKey, apiObj)
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

	err = setPlanValue(r.FormValue("rl-unauth-global-daily"), apiObj.Spec.SetUnAuthGlobalDaily)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-global-monthly"), apiObj.Spec.SetUnAuthGlobalMonthly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-global-eternity"), apiObj.Spec.SetUnAuthGlobalEternity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-remoteIP-daily"), apiObj.Spec.SetUnAuthRemoteIPDaily)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-remoteIP-monthly"), apiObj.Spec.SetUnAuthRemoteIPMonthly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-remoteIP-eternity"), apiObj.Spec.SetUnAuthRemoteIPEternity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, pathItem := range doc.Paths {
		for _, operation := range pathItem.Operations() {
			err := setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-unauth-%s-daily", operation.OperationID)),
				operation.OperationID,
				apiObj.Spec.SetUnAuthOperationDaily,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-unauth-%s-monthly", operation.OperationID)),
				operation.OperationID,
				apiObj.Spec.SetUnAuthOperationMonthly,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-unauth-%s-eternity", operation.OperationID)),
				operation.OperationID,
				apiObj.Spec.SetUnAuthOperationEternity,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}

	err = a.K8sClient.Update(context.Background(), apiObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
