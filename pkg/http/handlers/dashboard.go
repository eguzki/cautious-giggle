package handlers

import (
	"context"
	"html/template"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
)

type DashboardHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &DashboardHandler{}

func (a *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFS(giggletemplates.TemplatesFS, "dashboard.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiList := &gigglev1alpha1.ApiList{}
	err = a.K8sClient.List(context.Background(), apiList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := apiList.Items
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
