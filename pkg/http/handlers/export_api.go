package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	corev1 "k8s.io/api/core/v1"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/eguzki/cautious-giggle/pkg/istiogenerators"
	"github.com/eguzki/cautious-giggle/pkg/kuadrantgenerators"
)

type ExportAPIHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &ExportAPIHandler{}

func (a *ExportAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["api"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "url param api not found", http.StatusBadRequest)
		return
	}

	apiName := keys[0]
	apikey := client.ObjectKey{Name: apiName, Namespace: "default"}
	apiObj := &gigglev1alpha1.Api{}
	err := a.K8sClient.Get(context.Background(), apikey, apiObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if apiObj.Spec.Gateway == nil {
		http.Error(w, "API with empty gateway", http.StatusBadRequest)
		return
	}

	gatewayServiceKey := client.ObjectKey{Name: *apiObj.Spec.Gateway, Namespace: "default"}
	gatewayServiceObj := &corev1.Service{}
	err = a.K8sClient.Get(context.Background(), gatewayServiceKey, gatewayServiceObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gatewayLabels := []string{}
	for key, val := range gatewayServiceObj.Spec.Selector {
		gatewayLabels = append(gatewayLabels, fmt.Sprintf("%s=%s", key, val))
	}

	openapiLoader := openapi3.NewLoader()
	doc, err := openapiLoader.LoadFromData([]byte(apiObj.Spec.OAS))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = doc.Validate(openapiLoader.Context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vs, err := istiogenerators.VirtualService(doc, apiObj.Spec.ServiceName,
		"default", 80, []string{*apiObj.Spec.Gateway}, apiObj.Spec.PublicDomain, apiObj.Spec.PathMatchType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authzPolicy, err := istiogenerators.AuthorizationPolicy(doc, gatewayLabels, apiObj.Spec.PublicDomain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authConfig, err := kuadrantgenerators.AuthConfig(doc, apiObj.Spec.PublicDomain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	serializer := k8sJson.NewSerializerWithOptions(
		k8sJson.DefaultMetaFactory, nil, nil,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Strict: true,
		},
	)

	//w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Type", "application/x-yaml")

	frameWriter := k8sJson.YAMLFramer.NewFrameWriter(w)

	err = serializer.Encode(vs, frameWriter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = serializer.Encode(authzPolicy, frameWriter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = serializer.Encode(authConfig, frameWriter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
