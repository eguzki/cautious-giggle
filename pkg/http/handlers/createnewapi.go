package handlers

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglekuadrantiov1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/eguzki/cautious-giggle/pkg/utils"
	"github.com/getkin/kin-openapi/openapi3"
)

type CreateNewAPIHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &CreateNewAPIHandler{}

func (a *CreateNewAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	serviceName := r.FormValue("service")
	if serviceName == "" {
		http.Error(w, "form param service not found", http.StatusBadRequest)
		return
	}

	apiDescr := r.FormValue("description")
	if apiDescr == "" {
		http.Error(w, "form param description not found", http.StatusBadRequest)
		return
	}

	publicDomain := r.FormValue("publicdomain")
	if publicDomain == "" {
		http.Error(w, "form param publicdomain not found", http.StatusBadRequest)
		return
	}

	matchType := r.FormValue("matchtype")
	if matchType == "" {
		http.Error(w, "form param matchtype not found", http.StatusBadRequest)
		return
	}

	oasContent := r.FormValue("oas")

	if oasContent == "" {
		serviceKey := client.ObjectKey{Name: serviceName, Namespace: "default"}
		serviceObj := &corev1.Service{}
		err = a.K8sClient.Get(context.Background(), serviceKey, serviceObj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		oasConfigmapName, ok := serviceObj.Annotations[utils.KuadrantDiscoveryAnnotationOASConfigMap]
		if !ok {
			http.Error(w, "service does not specify OAS configmap", http.StatusInternalServerError)
			return
		}

		oasConfigmap := &corev1.ConfigMap{}
		oasConfigMapKey := client.ObjectKey{Name: oasConfigmapName, Namespace: serviceObj.Namespace}
		err = a.K8sClient.Get(context.Background(), oasConfigMapKey, oasConfigmap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		oasContent, ok = oasConfigmap.Data["openapi.yaml"]
		if !ok {
			http.Error(w, "oas configmap is missing the openapi.yaml entry", http.StatusInternalServerError)
			return
		}
	}

	desiredAPIObj := &gigglekuadrantiov1alpha1.Api{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Api",
			APIVersion: gigglekuadrantiov1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: "default",
		},
		Spec: gigglekuadrantiov1alpha1.ApiSpec{
			Description:   apiDescr,
			ServiceName:   serviceName,
			PublicDomain:  publicDomain,
			OAS:           oasContent,
			PathMatchType: matchType,
		},
	}

	// Fill plan
	err = setPlanValue(r.FormValue("rl-unauth-global-daily"), desiredAPIObj.Spec.SetUnAuthGlobalDaily)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-global-monthly"), desiredAPIObj.Spec.SetUnAuthGlobalMonthly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-global-yearly"), desiredAPIObj.Spec.SetUnAuthGlobalYearly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-remoteIP-daily"), desiredAPIObj.Spec.SetUnAuthRemoteIPDaily)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-remoteIP-monthly"), desiredAPIObj.Spec.SetUnAuthRemoteIPMonthly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = setPlanValue(r.FormValue("rl-unauth-remoteIP-yearly"), desiredAPIObj.Spec.SetUnAuthRemoteIPYearly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	openapiLoader := openapi3.NewLoader()
	doc, err := openapiLoader.LoadFromData([]byte(desiredAPIObj.Spec.OAS))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = doc.Validate(openapiLoader.Context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, pathItem := range doc.Paths {
		for _, operation := range pathItem.Operations() {
			err := setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-unauth-%s-daily", operation.OperationID)),
				operation.OperationID,
				desiredAPIObj.Spec.SetUnAuthOperationDaily,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-unauth-%s-monthly", operation.OperationID)),
				operation.OperationID,
				desiredAPIObj.Spec.SetUnAuthOperationMonthly,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = setPlanOperationValue(
				r.FormValue(fmt.Sprintf("rl-unauth-%s-yearly", operation.OperationID)),
				operation.OperationID,
				desiredAPIObj.Spec.SetUnAuthOperationYearly,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}

	apiKey := client.ObjectKey{Name: desiredAPIObj.Name, Namespace: desiredAPIObj.Namespace}
	existingAPIObj := &gigglekuadrantiov1alpha1.Api{}
	err = a.K8sClient.Get(context.Background(), apiKey, existingAPIObj)
	if err != nil {
		if !errors.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = a.K8sClient.Create(context.Background(), desiredAPIObj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	if !reflect.DeepEqual(existingAPIObj.Spec, desiredAPIObj.Spec) {
		existingAPIObj.Spec = desiredAPIObj.Spec
		err = a.K8sClient.Update(context.Background(), existingAPIObj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}
