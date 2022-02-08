package handlers

import (
	"context"
	"net/http"

	istionetworkingapi "istio.io/api/networking/v1beta1"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eguzki/cautious-giggle/pkg/utils"
)

type CreateGatewaysHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &CreateGatewaysHandler{}

func (a *CreateGatewaysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gwName := r.FormValue("name")
	if gwName == "" {
		http.Error(w, "form param name not found", http.StatusBadRequest)
		return
	}

	labelKey1 := r.FormValue("labelkey1")
	if labelKey1 == "" {
		http.Error(w, "form param labelkey1 not found", http.StatusBadRequest)
		return
	}

	labelValue1 := r.FormValue("labelvalue1")
	if labelValue1 == "" {
		http.Error(w, "form param labelvalue1 not found", http.StatusBadRequest)
		return
	}

	gwDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
			Labels:    map[string]string{utils.KuadrantGatewayLabel: "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{labelKey1: labelValue1},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						// Set a unique label for the gateway. This is required to ensure Gateways can select this workload
						labelKey1: labelValue1,
						// Enable gateway injection. If connecting to a revisioned control plane, replace with "istio.io/rev: revision-name"
						"sidecar.istio.io/inject": "true",
					},
					Annotations: map[string]string{
						// Select the gateway injection template (rather than the default sidecar template)
						"inject.istio.io/templates": "gateway",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						v1.Container{
							Name: "istio-proxy",
							// The image will automatically update each time the pod starts.
							Image: "auto",
						},
					},
				},
			},
		},
	}

	if err := a.K8sClient.Create(context.Background(), gwDeployment); utils.IgnoreAlreadyExists(err) != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gwService := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
			Labels:    map[string]string{utils.KuadrantGatewayLabel: "true"},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Name: "http", Port: 80},
				{Name: "https", Port: 443},
			},
			Selector: map[string]string{labelKey1: labelValue1},
		},
	}

	if err := a.K8sClient.Create(context.Background(), gwService); utils.IgnoreAlreadyExists(err) != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gateway := &istionetworking.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
			Labels:    map[string]string{utils.KuadrantGatewayLabel: "true"},
		},
		Spec: istionetworkingapi.Gateway{
			Selector: map[string]string{labelKey1: labelValue1},
			Servers: []*istionetworkingapi.Server{
				&istionetworkingapi.Server{
					Hosts: []string{"*"},
					Port: &istionetworkingapi.Port{
						Number:   80,
						Name:     "http",
						Protocol: "HTTP",
					},
				},
			},
		},
	}

	if err := a.K8sClient.Create(context.Background(), gateway); utils.IgnoreAlreadyExists(err) != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/gateways", http.StatusFound)
}
