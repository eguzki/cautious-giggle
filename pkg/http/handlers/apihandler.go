package handlers

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIHandler struct {
	K8sClient client.Client
}

func NewAPIHandler(k8sClient client.Client) *APIHandler {
	return &APIHandler{
		K8sClient: k8sClient,
	}
}
