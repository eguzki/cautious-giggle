package kuadrantgenerators

import (
	"github.com/getkin/kin-openapi/openapi3"
	kuadrantv1alpha1 "github.com/kuadrant/kuadrant-controller/apis/apim/v1alpha1"
	limitadorv1alpha1 "github.com/kuadrant/limitador-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/eguzki/cautious-giggle/pkg/istiogenerators"
	"github.com/eguzki/cautious-giggle/pkg/utils"
)

func RateLimitPolicy(doc *openapi3.T, api *gigglev1alpha1.Api) (*kuadrantv1alpha1.RateLimitPolicy, error) {
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
			RateLimits: globalActions,
			Routes:     routes,
			Limits:     limits,
		},
	}
	return rlp, nil
}

func generateRLPRoutes(doc *openapi3.T, api *gigglev1alpha1.Api) []kuadrantv1alpha1.Route {
	routes := make([]kuadrantv1alpha1.Route, 0)
	for _, pathItem := range doc.Paths {
		for _, operation := range pathItem.Operations() {
			route := kuadrantv1alpha1.Route{
				Name: operation.OperationID,
			}
			// Unauth
			if !api.Spec.GetUnAuthRateLimit().GetOperation(operation.OperationID).IsEmpty() {
				pathMethodRateLimit := istiogenerators.PathMethodRateLimit(kuadrantv1alpha1.RateLimitStagePREAUTH)
				route.RateLimits = append(route.RateLimits, pathMethodRateLimit)
			}
			// Auth
			if api.Spec.HasAnyRateLimitOnOperation(operation.OperationID) {
				pathMethodRateLimit := istiogenerators.PathMethodRateLimit(kuadrantv1alpha1.RateLimitStagePOSTAUTH)
				route.RateLimits = append(route.RateLimits, pathMethodRateLimit)
			}
			if len(route.RateLimits) > 0 {
				routes = append(routes, route)
			}
		}
	}

	return routes
}

func generateRLPLimits(doc *openapi3.T, api *gigglev1alpha1.Api) []limitadorv1alpha1.RateLimitSpec {
	return nil
}

func generateRLPGlobalActions(doc *openapi3.T, api *gigglev1alpha1.Api) []*kuadrantv1alpha1.RateLimit {
	tmpAPIStr := "api"
	preRateLimit := &kuadrantv1alpha1.RateLimit{
		Stage: kuadrantv1alpha1.RateLimitStagePREAUTH,
		Actions: []*kuadrantv1alpha1.ActionSpecifier{
			// Generic descriptor enty
			// ("api", "API_NAME")
			// TODO(eastizle): should only be added if there is at least one rate limit set
			&kuadrantv1alpha1.ActionSpecifier{
				GenericKey: &kuadrantv1alpha1.GenericKeySpec{
					DescriptorKey:   &tmpAPIStr,
					DescriptorValue: api.Name,
				},
			},
		},
	}

	postRateLimit := &kuadrantv1alpha1.RateLimit{
		Stage: kuadrantv1alpha1.RateLimitStagePOSTAUTH,
		Actions: []*kuadrantv1alpha1.ActionSpecifier{
			// Generic descriptor enty
			// ("api", "API_NAME")
			// TODO(eastizle): should only be added if there is at least one rate limit set
			&kuadrantv1alpha1.ActionSpecifier{
				GenericKey: &kuadrantv1alpha1.GenericKeySpec{
					DescriptorKey:   &tmpAPIStr,
					DescriptorValue: api.Name,
				},
			},
		},
	}

	if !api.Spec.GetUnAuthRateLimit().GetRemoteIP().IsEmpty() {
		// UnAuth RemoteIP
		preRateLimit.Actions = append(preRateLimit.Actions, &kuadrantv1alpha1.ActionSpecifier{
			RemoteAddress: &kuadrantv1alpha1.RemoteAddressSpec{},
		})
	}

	if api.Spec.HasAnyAuthRateLimit() {
		// auth user-id
		//- metadata:
		//    descriptor_key: "user_id"
		//    metadata_key:
		//      key: "envoy.filters.http.ext_authz"
		//      path:
		//        - key: "ext_auth_data"
		//        - key: "user_id"
		postRateLimit.Actions = append(postRateLimit.Actions, &kuadrantv1alpha1.ActionSpecifier{
			Metadata: &kuadrantv1alpha1.MetadataSpec{
				DescriptorKey: "user_id",
				MetadataKey: kuadrantv1alpha1.MetadataKeySpec{
					Key: "envoy.filters.http.ext_authz",
					Path: []kuadrantv1alpha1.MetadataPathSegment{
						{
							Key: "ext_auth_data",
						},
						{
							Key: "user_id",
						},
					},
				},
			},
		})
	}

	// no auth RemoteIP action

	return []*kuadrantv1alpha1.RateLimit{preRateLimit, postRateLimit}
}
