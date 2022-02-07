package handlers

import (
	"context"
	"html/template"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
)

type UsersHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &UsersHandler{}

func (a *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userList := &gigglev1alpha1.UserList{}
	err := a.K8sClient.List(context.Background(), userList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "users.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, userList.Items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
