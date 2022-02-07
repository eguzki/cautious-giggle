package handlers

import (
	"context"
	"fmt"
	"net/http"

	gigglekuadrantiov1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CreateUserHandler struct {
	K8sClient client.Client
}

var _ http.Handler = &CreateUserHandler{}

func (a *CreateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userName := r.FormValue("name")
	if userName == "" {
		http.Error(w, "form param name not found", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("id")
	if userID == "" {
		http.Error(w, "form param id not found", http.StatusBadRequest)
		return
	}

	userKey := client.ObjectKey{Name: userID, Namespace: "default"}
	existingUserObj := &gigglekuadrantiov1alpha1.User{}
	err = a.K8sClient.Get(context.Background(), userKey, existingUserObj)
	if err != nil && !errors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil && errors.IsNotFound(err) {
		user := &gigglekuadrantiov1alpha1.User{
			TypeMeta: metav1.TypeMeta{
				Kind:       "User",
				APIVersion: gigglekuadrantiov1alpha1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      userID,
				Namespace: "default",
			},
			Spec: gigglekuadrantiov1alpha1.UserSpec{
				LongName: userName,
			},
		}
		err = a.K8sClient.Create(context.Background(), user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	apiList := &gigglev1alpha1.ApiList{}
	err = a.K8sClient.List(context.Background(), apiList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for apiIdx := range apiList.Items {
		planSelected := r.FormValue(fmt.Sprintf("%splan", apiList.Items[apiIdx].Name))
		if planSelected == "" || planSelected == "-" {
			continue
		}

		for planID := range apiList.Items[apiIdx].Spec.Plans {
			if planSelected == planID {
				if apiList.Items[apiIdx].Spec.UserPlan == nil {
					apiList.Items[apiIdx].Spec.UserPlan = map[string]string{}
				}
				apiList.Items[apiIdx].Spec.UserPlan[userID] = planID

				err = a.K8sClient.Update(context.Background(), &apiList.Items[apiIdx])
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	http.Redirect(w, r, "/users", http.StatusFound)
}
