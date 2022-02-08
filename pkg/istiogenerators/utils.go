package istiogenerators

import (
	"github.com/getkin/kin-openapi/openapi3"
	istioapi "istio.io/api/networking/v1beta1"
	istiosecurityapi "istio.io/api/security/v1beta1"

	"github.com/eguzki/cautious-giggle/pkg/utils"
)

func HTTPRoutesFromOpenAPI(oasDoc *openapi3.T, destination *istioapi.Destination) []*istioapi.HTTPRoute {
	httpRoutes := []*istioapi.HTTPRoute{}

	// Path based routing
	for path, pathItem := range oasDoc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			httpRoute := &istioapi.HTTPRoute{
				// TODO(eastizle): OperationID can be null, fallback to some custom name
				Name: operation.OperationID,
				Match: []*istioapi.HTTPMatchRequest{
					{
						Uri: &istioapi.StringMatch{
							MatchType: &istioapi.StringMatch_Exact{Exact: path},
						},
						Method: &istioapi.StringMatch{
							MatchType: &istioapi.StringMatch_Exact{Exact: opVerb},
						},
					},
				},
				Route: []*istioapi.HTTPRouteDestination{{Destination: destination}},
			}
			httpRoutes = append(httpRoutes, httpRoute)
		}
	}

	return httpRoutes
}

func AuthorizationPolicyRulesFromOpenAPI(oasDoc *openapi3.T, publicDomain string) []*istiosecurityapi.Rule {
	rules := []*istiosecurityapi.Rule{}

	for path, pathItem := range oasDoc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			secReqsP := utils.OpenAPIOperationSecRequirements(oasDoc, operation)

			if secReqsP == nil || len(*secReqsP) == 0 {
				continue
			}

			// there is at least one sec requirement for this operation,
			// add the operation to authorization policy rules
			rule := &istiosecurityapi.Rule{
				To: []*istiosecurityapi.Rule_To{
					{
						Operation: &istiosecurityapi.Operation{
							Hosts:   []string{publicDomain},
							Methods: []string{opVerb},
							Paths:   []string{path},
						},
					},
				},
			}

			rules = append(rules, rule)
		}
	}
	return rules
}
