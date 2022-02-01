package handlers

import (
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &APIHandler{}

func (a *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
