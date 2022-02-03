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

type Gateway struct {
	Name   string
	Labels map[string]string
}

type GatewaysHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &GatewaysHandler{}

func (a *GatewaysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceList := &corev1.ServiceList{}
	err := a.K8sClient.List(context.Background(), serviceList, client.HasLabels{utils.KuadrantGatewayLabel})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := []Gateway{}
	for idx := range serviceList.Items {
		gateway := Gateway{
			Name:   serviceList.Items[idx].Name,
			Labels: serviceList.Items[idx].Spec.Selector,
		}
		data = append(data, gateway)
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "gateways.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
