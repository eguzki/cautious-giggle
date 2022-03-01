package kuadrantgenerators

import (
	"github.com/getkin/kin-openapi/openapi3"
	authorinov1beta1 "github.com/kuadrant/authorino/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/eguzki/cautious-giggle/pkg/utils"
)

func AuthConfig(doc *openapi3.T, publicHost string) (*authorinov1beta1.AuthConfig, error) {
	objectName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		return nil, err
	}

	identityList, err := AuthConfigIdentitiesFromOpenAPI(doc)
	if err != nil {
		return nil, err
	}

	authConfig := &authorinov1beta1.AuthConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AuthConfig",
			APIVersion: "authorino.kuadrant.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: objectName,
		},
		Spec: authorinov1beta1.AuthConfigSpec{
			Hosts:         []string{publicHost},
			Identity:      identityList,
			Metadata:      nil,
			Authorization: nil,
			Response:      AuthConfigResponses(),
			Patterns:      nil,
			Conditions:    nil,
		},
	}
	return authConfig, nil
}
