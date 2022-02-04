package handlers

import (
	"context"
	"fmt"
	"net/http"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
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
		http.Error(w, "form param api not found", http.StatusBadRequest)
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

	plan := gigglev1alpha1.ApiPlan{
		Description: description,
	}

	// Fill plan

	if apiObj.Spec.Plans == nil {
		apiObj.Spec.Plans = map[string]gigglev1alpha1.ApiPlan{}
	}

	apiObj.Spec.Plans[name] = plan
	err = a.K8sClient.Update(context.Background(), apiObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/api?id=%s", apiName), http.StatusFound)
}
