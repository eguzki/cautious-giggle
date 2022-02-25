package istiogenerators

import (
	"github.com/getkin/kin-openapi/openapi3"
	istiosecurityapi "istio.io/api/security/v1beta1"
	istiotypeapi "istio.io/api/type/v1beta1"
	istiosecurity "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/eguzki/cautious-giggle/pkg/utils"
)

func AuthorizationPolicy(doc *openapi3.T, gatewayLabels map[string]string, publicHost string) (*istiosecurity.AuthorizationPolicy, error) {

	objectName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		return nil, err
	}

	authPolicy := &istiosecurity.AuthorizationPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AuthorizationPolicy",
			APIVersion: "security.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			// Missing namespace
			Name: objectName,
		},
		Spec: istiosecurityapi.AuthorizationPolicy{
			Selector: &istiotypeapi.WorkloadSelector{
				MatchLabels: gatewayLabels,
			},
			Rules:  AuthorizationPolicyRulesFromOpenAPI(doc, publicHost),
			Action: istiosecurityapi.AuthorizationPolicy_CUSTOM,
			ActionDetail: &istiosecurityapi.AuthorizationPolicy_Provider{
				Provider: &istiosecurityapi.AuthorizationPolicy_ExtensionProvider{
					Name: "kuadrant-authorization",
				},
			},
		},
	}

	return authPolicy, nil
}
