package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CreateRateLimitPlanHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &CreateRateLimitPlanHandler{}

func (a *CreateRateLimitPlanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "form param description not found", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "form param name not found", http.StatusBadRequest)
		return
	}

	if _, ok := apiObj.Spec.Plans[name]; ok {
		http.Error(w, "rate limit plan exists", http.StatusBadRequest)
		return
	}

	plan := &gigglev1alpha1.ApiPlan{
		Description: description,
	}

	// Fill plan
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

	err = setPlanValue(r.FormValue("rl-auth-global-daily"), plan.SetAuthGlobalDaily)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-auth-global-monthly"), plan.SetAuthGlobalMonthly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-auth-global-yearly"), plan.SetAuthGlobalYearly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, pathItem := range doc.Paths {
		for _, operation := range pathItem.Operations() {
			err := setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-auth-%s-daily", operation.OperationID)),
				operation.OperationID,
				plan.SetAuthOperationDaily,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-auth-%s-monthly", operation.OperationID)),
				operation.OperationID,
				plan.SetAuthOperationMonthly,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-auth-%s-yearly", operation.OperationID)),
				operation.OperationID,
				plan.SetAuthOperationYearly,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}

	if apiObj.Spec.Plans == nil {
		apiObj.Spec.Plans = map[string]*gigglev1alpha1.ApiPlan{}
	}

	apiObj.Spec.Plans[name] = plan
	err = a.K8sClient.Update(context.Background(), apiObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/api?id=%s", apiName), http.StatusFound)
}

func setPlanValue(valStr string, cb func(val int32)) error {
	if valStr != "" && valStr != "0" {
		tmp64, err := strconv.ParseInt(valStr, 10, 32)
		if err != nil {
			return err
		}
		cb(int32(tmp64))
	}
	return nil
}

func setPlanOperationValue(valStr, operationID string, cb func(val int32, operationID string)) error {
	if valStr != "" && valStr != "0" {
		tmp64, err := strconv.ParseInt(valStr, 10, 32)
		if err != nil {
			return err
		}
		cb(int32(tmp64), operationID)
	}
	return nil
}
