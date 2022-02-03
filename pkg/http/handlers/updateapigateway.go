package handlers

import (
	"context"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
)

type UpdateAPIGatewayHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &UpdateAPIGatewayHandler{}

func (a *UpdateAPIGatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	gw := r.FormValue("gw")
	if gw == "" {
		http.Error(w, "form param gw not found", http.StatusBadRequest)
		return
	}

	apiObj := &gigglev1alpha1.Api{}
	apiKey := client.ObjectKey{Name: apiName, Namespace: "default"}
	err = a.K8sClient.Get(context.Background(), apiKey, apiObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if apiObj.Spec.Gateway == nil || *apiObj.Spec.Gateway != gw {
		apiObj.Spec.Gateway = &gw
		err = a.K8sClient.Update(context.Background(), apiObj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
