package kuadrantgenerators

import (
	"github.com/getkin/kin-openapi/openapi3"
	kuadrantv1alpha1 "github.com/kuadrant/kuadrant-controller/apis/apim/v1alpha1"
	limitadorv1alpha1 "github.com/kuadrant/limitador-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/eguzki/cautious-giggle/pkg/utils"
)

func RateLimitPolicy(doc *openapi3.T, vsName string, api *gigglev1alpha1.Api) (*kuadrantv1alpha1.RateLimitPolicy, error) {
	objectName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		return nil, err
	}

	routes := generateRLPRoutes(doc, api)
	limits := generateRLPLimits(doc, api)
	globalActions := generateRLPGlobalActions(doc, api)

	rlp := &kuadrantv1alpha1.RateLimitPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RateLimitPolicy",
			APIVersion: "apim.kuadrant.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: objectName,
		},
		Spec: kuadrantv1alpha1.RateLimitPolicySpec{
			NetworkingRef: []kuadrantv1alpha1.NetworkingRef{
				{
					Type: kuadrantv1alpha1.NetworkingRefType_VS,
					Name: vsName,
				},
			},
			Actions: globalActions,
			Routes:  routes,
			Limits:  limits,
		},
	}
	return rlp, nil
}

func generateRLPRoutes(doc *openapi3.T, api *gigglev1alpha1.Api) []kuadrantv1alpha1.Route {
	return nil
}

func generateRLPLimits(doc *openapi3.T, api *gigglev1alpha1.Api) []limitadorv1alpha1.RateLimitSpec {
	return nil
}

func generateRLPGlobalActions(doc *openapi3.T, api *gigglev1alpha1.Api) []*kuadrantv1alpha1.Action_Specifier {
	// UnAuth Global
	// UnAuth RemoteIP
	// Auth Global

	return nil
}
