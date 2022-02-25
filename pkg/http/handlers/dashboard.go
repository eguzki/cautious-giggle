package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
)

type DashboardAPI struct {
	Name        string
	Description string
	Action      string
	GwLinked    bool
}

type DashboardData struct {
	APIs []*DashboardAPI
}

type DashboardHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &DashboardHandler{}

func (a *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiList := &gigglev1alpha1.ApiList{}
	err := a.K8sClient.List(context.Background(), apiList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &DashboardData{}

	for idx := range apiList.Items {
		api := &DashboardAPI{
			Name:        apiList.Items[idx].Name,
			Description: apiList.Items[idx].Spec.Description,
			Action:      fmt.Sprintf("kubectl apply -f http://127.0.0.1:8082/export?api=%s", apiList.Items[idx].Name),
			GwLinked:    false,
		}
		data.APIs = append(data.APIs, api)
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "dashboard.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
