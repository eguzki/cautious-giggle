package handlers

import (
	"context"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/eguzki/cautious-giggle/pkg/istiogenerators"
	"github.com/eguzki/cautious-giggle/pkg/kuadrantgenerators"
	"github.com/eguzki/cautious-giggle/pkg/utils"
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

	gatewayLabels := map[string]string{
		"istio": "kuadrant-system",
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

	rateLimiPolicy, err := kuadrantgenerators.RateLimitPolicy(doc, apiObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vs, err := istiogenerators.VirtualService(doc, apiObj.Spec.ServiceName,
		"default", 80, []string{"kuadrant-system/kuadrant-gateway"}, rateLimiPolicy.Name,
		apiObj.Spec.PublicDomain, apiObj.Spec.PathMatchType)
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

	workloadName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	apiKeySecrets := kuadrantgenerators.APIKeySecrets(workloadName, apiObj)

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

	err = serializer.Encode(rateLimiPolicy, frameWriter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for idx := range apiKeySecrets {
		err = serializer.Encode(apiKeySecrets[idx], frameWriter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

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
