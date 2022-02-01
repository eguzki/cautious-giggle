package handlers

import (
	"context"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglekuadrantiov1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
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

	gwDescr := r.FormValue("description")
	if gwDescr == "" {
		http.Error(w, "form param description not found", http.StatusBadRequest)
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

	labelKey2 := r.FormValue("labelkey2")
	if labelKey2 == "" {
		http.Error(w, "form param labelkey2 not found", http.StatusBadRequest)
		return
	}

	labelValue2 := r.FormValue("labelvalue2")
	if labelValue2 == "" {
		http.Error(w, "form param labelvalue2 not found", http.StatusBadRequest)
		return
	}

	desiredGwObj := &gigglekuadrantiov1alpha1.Gateway{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Gateway",
			APIVersion: gigglekuadrantiov1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gwName,
			Namespace: "default",
		},
		Spec: gigglekuadrantiov1alpha1.GatewaySpec{
			Description: gwDescr,
			Labels: map[string]string{
				labelKey1: labelValue1,
				labelKey2: labelValue2,
			},
		},
	}

	if err := a.K8sClient.Create(context.Background(), desiredGwObj); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/gateways", http.StatusFound)
}
