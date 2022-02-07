package handlers

import (
	"context"
	"html/template"
	"net/http"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	giggletemplates "github.com/eguzki/cautious-giggle/pkg/http/templates"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NewUserHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &NewUserHandler{}

type NewUserPlanInfo struct {
	Name string
}

type NewUserAPIInfo struct {
	APIName string
	Plans   []*NewUserPlanInfo
}

type NewUserData struct {
	APIs []*NewUserAPIInfo
}

func (a *NewUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiList := &gigglev1alpha1.ApiList{}
	err := a.K8sClient.List(context.Background(), apiList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := &NewUserData{}

	for idx := range apiList.Items {
		apiInfo := &NewUserAPIInfo{
			APIName: apiList.Items[idx].Name,
		}

		for planID := range apiList.Items[idx].Spec.Plans {
			apiInfo.Plans = append(apiInfo.Plans, &NewUserPlanInfo{
				Name: planID,
			})
		}

		data.APIs = append(data.APIs, apiInfo)
	}

	t, err := template.ParseFS(giggletemplates.TemplatesFS, "newuser.html.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
