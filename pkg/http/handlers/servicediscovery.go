package handlers

import (
	"context"
	"html/template"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
	"github.com/eguzki/cautious-giggle/pkg/utils"
)

type ServiceDiscoveryHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &ServiceDiscoveryHandler{}

func (a *ServiceDiscoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	serviceList := &corev1.ServiceList{}
	labels := []string{utils.KuadrantDiscoveryLabel}
	err := a.K8sClient.List(context.Background(), serviceList, client.HasLabels(labels))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "servicediscovery.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, serviceList.Items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
