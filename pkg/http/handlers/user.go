package handlers

import (
	"context"
	"html/template"
	"net/http"

	gigglekuadrantiov1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UserAPIInfo struct {
	APIName  string
	PlanName string
}

type UserData struct {
	APIs []*UserAPIInfo
	Name string
	ID   string
}

type UserHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &UserHandler{}

func (a *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["id"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "url param id not found", http.StatusBadRequest)
		return
	}

	userID := keys[0]

	userKey := client.ObjectKey{Name: userID, Namespace: "default"}
	userObj := &gigglekuadrantiov1alpha1.User{}
	err := a.K8sClient.Get(context.Background(), userKey, userObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiList := &gigglev1alpha1.ApiList{}
	err = a.K8sClient.List(context.Background(), apiList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &UserData{
		Name: userObj.Spec.LongName,
		ID:   userObj.Name,
	}

	for idx := range apiList.Items {
		apiInfo := &UserAPIInfo{
			APIName: apiList.Items[idx].Name,
		}

		if planName, ok := apiList.Items[idx].Spec.UserPlan[userID]; ok {
			apiInfo.PlanName = planName
		} else {
			apiInfo.PlanName = "Not assigned"
		}

		data.APIs = append(data.APIs, apiInfo)
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "user.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
