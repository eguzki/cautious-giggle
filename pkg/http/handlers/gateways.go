package handlers

import (
	"context"
	"html/template"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglekuadrantiov1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
)

type GatewaysHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &GatewaysHandler{}

func (a *GatewaysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gatewayList := &gigglekuadrantiov1alpha1.GatewayList{}
	err := a.K8sClient.List(context.Background(), gatewayList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "gateways.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, gatewayList.Items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
