package kuadrantgenerators

import (
	"fmt"

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
	globalActions := generateRLPGlobalActions(api)

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
	limits := make([]limitadorv1alpha1.RateLimitSpec, 0)

	tmp := generateGlobalUnAuthLimits(api)
	limits = append(limits, tmp...)

	tmp = generateRemoteIPUnAuthLimits(api)
	limits = append(limits, tmp...)

	tmp = generateUnAuthRouteLimits(doc, api)
	limits = append(limits, tmp...)

	tmp = generateGlobalAuthLimits(api)
	limits = append(limits, tmp...)

	tmp = generateAuthRouteLimits(doc, api)
	limits = append(limits, tmp...)

	return limits
}

func generateRLPGlobalActions(api *gigglev1alpha1.Api) []*kuadrantv1alpha1.RateLimit {
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
		//    descriptor_key: "user-id"
		//    metadata_key:
		//      key: "envoy.filters.http.ext_authz"
		//      path:
		//        - key: "ext_auth_data"
		//        - key: "user-id"
		postRateLimit.Actions = append(postRateLimit.Actions, &kuadrantv1alpha1.ActionSpecifier{
			Metadata: &kuadrantv1alpha1.MetadataSpec{
				DescriptorKey: "user-id",
				MetadataKey: kuadrantv1alpha1.MetadataKeySpec{
					Key: "envoy.filters.http.ext_authz",
					Path: []kuadrantv1alpha1.MetadataPathSegment{
						{
							Key: "ext_auth_data",
						},
						{
							Key: "user-id",
						},
					},
				},
			},
		})
	}

	// no auth RemoteIP action

	return []*kuadrantv1alpha1.RateLimit{preRateLimit, postRateLimit}
}

func generateGlobalUnAuthLimits(api *gigglev1alpha1.Api) []limitadorv1alpha1.RateLimitSpec {
	limits := make([]limitadorv1alpha1.RateLimitSpec, 0)

	globalUnAuth := api.Spec.GetUnAuthRateLimit().GetGlobal()
	if globalUnAuth.Daily != nil {
		limits = append(limits, limitadorv1alpha1.RateLimitSpec{
			Namespace:  "preauth",
			Conditions: []string{fmt.Sprintf("api == %s", api.Name)},
			Variables:  []string{},
			MaxValue:   int(*globalUnAuth.Daily),
			Seconds:    3600 * 24,
		})
	}

	if globalUnAuth.Monthly != nil {
		limits = append(limits, limitadorv1alpha1.RateLimitSpec{
			Namespace:  "preauth",
			Conditions: []string{fmt.Sprintf("api == %s", api.Name)},
			Variables:  []string{},
			MaxValue:   int(*globalUnAuth.Monthly),
			Seconds:    3600 * 24 * 30,
		})
	}

	if globalUnAuth.Yearly != nil {
		limits = append(limits, limitadorv1alpha1.RateLimitSpec{
			Namespace:  "preauth",
			Conditions: []string{fmt.Sprintf("api == %s", api.Name)},
			Variables:  []string{},
			MaxValue:   int(*globalUnAuth.Yearly),
			Seconds:    3600 * 24 * 365,
		})
	}

	return limits
}

func generateRemoteIPUnAuthLimits(api *gigglev1alpha1.Api) []limitadorv1alpha1.RateLimitSpec {
	limits := make([]limitadorv1alpha1.RateLimitSpec, 0)

	remoteIPUnAuth := api.Spec.GetUnAuthRateLimit().GetRemoteIP()
	if remoteIPUnAuth.Daily != nil {
		limits = append(limits, limitadorv1alpha1.RateLimitSpec{
			Namespace:  "preauth",
			Conditions: []string{fmt.Sprintf("api == %s", api.Name)},
			Variables:  []string{"remote_address"},
			MaxValue:   int(*remoteIPUnAuth.Daily),
			Seconds:    3600 * 24,
		})
	}

	if remoteIPUnAuth.Monthly != nil {
		limits = append(limits, limitadorv1alpha1.RateLimitSpec{
			Namespace:  "preauth",
			Conditions: []string{fmt.Sprintf("api == %s", api.Name)},
			Variables:  []string{"remote_address"},
			MaxValue:   int(*remoteIPUnAuth.Monthly),
			Seconds:    3600 * 24 * 30,
		})
	}

	if remoteIPUnAuth.Yearly != nil {
		limits = append(limits, limitadorv1alpha1.RateLimitSpec{
			Namespace:  "preauth",
			Conditions: []string{fmt.Sprintf("api == %s", api.Name)},
			Variables:  []string{"remote_address"},
			MaxValue:   int(*remoteIPUnAuth.Yearly),
			Seconds:    3600 * 24 * 365,
		})
	}

	return limits
}

func generateUnAuthRouteLimits(doc *openapi3.T, api *gigglev1alpha1.Api) []limitadorv1alpha1.RateLimitSpec {
	limits := make([]limitadorv1alpha1.RateLimitSpec, 0)

	for path, pathItem := range doc.Paths {
		for opMethod, operation := range pathItem.Operations() {
			rateLimitConf := api.Spec.GetUnAuthRateLimit().GetOperation(operation.OperationID)
			if rateLimitConf.Daily != nil {
				limits = append(limits, limitadorv1alpha1.RateLimitSpec{
					Namespace: "preauth",
					Conditions: []string{
						fmt.Sprintf("api == %s", api.Name),
						fmt.Sprintf("method == %s", opMethod),
						fmt.Sprintf("path == %s", path),
					},
					Variables: []string{},
					MaxValue:  int(*rateLimitConf.Daily),
					Seconds:   3600 * 24,
				})
			}

			if rateLimitConf.Monthly != nil {
				limits = append(limits, limitadorv1alpha1.RateLimitSpec{
					Namespace: "preauth",
					Conditions: []string{
						fmt.Sprintf("api == %s", api.Name),
						fmt.Sprintf("method == %s", opMethod),
						fmt.Sprintf("path == %s", path),
					},
					Variables: []string{},
					MaxValue:  int(*rateLimitConf.Monthly),
					Seconds:   3600 * 24 * 30,
				})
			}

			if rateLimitConf.Yearly != nil {
				limits = append(limits, limitadorv1alpha1.RateLimitSpec{
					Namespace: "preauth",
					Conditions: []string{
						fmt.Sprintf("api == %s", api.Name),
						fmt.Sprintf("method == %s", opMethod),
						fmt.Sprintf("path == %s", path),
					},
					Variables: []string{},
					MaxValue:  int(*rateLimitConf.Yearly),
					Seconds:   3600 * 24 * 365,
				})
			}
		}
	}

	return limits
}

func generateGlobalAuthLimits(api *gigglev1alpha1.Api) []limitadorv1alpha1.RateLimitSpec {
	limits := make([]limitadorv1alpha1.RateLimitSpec, 0)

	for userID, userInfo := range api.Spec.Users {
		if userInfo.Plan == nil {
			continue
		}
		apiPlan := api.Spec.Plans[*userInfo.Plan]
		if apiPlan == nil {
			panic(fmt.Sprintf("plan does not exist %s", *userInfo.Plan))
		}
		globalAuth := apiPlan.GetGlobal()
		if globalAuth.Daily != nil {
			limits = append(limits, limitadorv1alpha1.RateLimitSpec{
				Namespace: "postauth",
				Conditions: []string{
					fmt.Sprintf("api == %s", api.Name),
					fmt.Sprintf("user-id == %s", userID),
				},
				Variables: []string{},
				MaxValue:  int(*globalAuth.Daily),
				Seconds:   3600 * 24,
			})
		}

		if globalAuth.Monthly != nil {
			limits = append(limits, limitadorv1alpha1.RateLimitSpec{
				Namespace: "postauth",
				Conditions: []string{
					fmt.Sprintf("api == %s", api.Name),
					fmt.Sprintf("user-id == %s", userID),
				},
				Variables: []string{},
				MaxValue:  int(*globalAuth.Monthly),
				Seconds:   3600 * 24 * 30,
			})
		}

		if globalAuth.Yearly != nil {
			limits = append(limits, limitadorv1alpha1.RateLimitSpec{
				Namespace: "postauth",
				Conditions: []string{
					fmt.Sprintf("api == %s", api.Name),
					fmt.Sprintf("user-id == %s", userID),
				},
				Variables: []string{},
				MaxValue:  int(*globalAuth.Yearly),
				Seconds:   3600 * 24 * 365,
			})
		}
	}

	return limits
}

func generateAuthRouteLimits(doc *openapi3.T, api *gigglev1alpha1.Api) []limitadorv1alpha1.RateLimitSpec {
	limits := make([]limitadorv1alpha1.RateLimitSpec, 0)

	for userID, userInfo := range api.Spec.Users {
		if userInfo.Plan == nil {
			continue
		}
		apiPlan := api.Spec.Plans[*userInfo.Plan]
		if apiPlan == nil {
			panic(fmt.Sprintf("plan does not exist %s", *userInfo.Plan))
		}

		for path, pathItem := range doc.Paths {
			for opMethod, operation := range pathItem.Operations() {
				rateLimitConf := apiPlan.GetOperation(operation.OperationID)
				if rateLimitConf.Daily != nil {
					limits = append(limits, limitadorv1alpha1.RateLimitSpec{
						Namespace: "postauth",
						Conditions: []string{
							fmt.Sprintf("api == %s", api.Name),
							fmt.Sprintf("method == %s", opMethod),
							fmt.Sprintf("path == %s", path),
							fmt.Sprintf("user-id == %s", userID),
						},
						Variables: []string{},
						MaxValue:  int(*rateLimitConf.Daily),
						Seconds:   3600 * 24,
					})
				}

				if rateLimitConf.Monthly != nil {
					limits = append(limits, limitadorv1alpha1.RateLimitSpec{
						Namespace: "postauth",
						Conditions: []string{
							fmt.Sprintf("api == %s", api.Name),
							fmt.Sprintf("method == %s", opMethod),
							fmt.Sprintf("path == %s", path),
							fmt.Sprintf("user-id == %s", userID),
						},
						Variables: []string{},
						MaxValue:  int(*rateLimitConf.Monthly),
						Seconds:   3600 * 24 * 30,
					})
				}

				if rateLimitConf.Yearly != nil {
					limits = append(limits, limitadorv1alpha1.RateLimitSpec{
						Namespace: "postauth",
						Conditions: []string{
							fmt.Sprintf("api == %s", api.Name),
							fmt.Sprintf("method == %s", opMethod),
							fmt.Sprintf("path == %s", path),
							fmt.Sprintf("user-id == %s", userID),
						},
						Variables: []string{},
						MaxValue:  int(*rateLimitConf.Yearly),
						Seconds:   3600 * 24 * 365,
					})
				}
			}
		}
	}

	return limits
}
