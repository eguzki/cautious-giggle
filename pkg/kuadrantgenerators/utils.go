package kuadrantgenerators

import (
	"fmt"
	"strings"

	gigglev1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/getkin/kin-openapi/openapi3"
	authorinov1beta1 "github.com/kuadrant/authorino/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/eguzki/cautious-giggle/pkg/utils"
)

func AuthConfigIdentitiesFromOpenAPI(oasDoc *openapi3.T) ([]*authorinov1beta1.Identity, error) {
	identities := []*authorinov1beta1.Identity{}

	workloadName, err := utils.K8sNameFromOpenAPITitle(oasDoc)
	if err != nil {
		return nil, err
	}

	for path, pathItem := range oasDoc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			secReqsP := utils.OpenAPIOperationSecRequirements(oasDoc, operation)

			if secReqsP == nil {
				continue
			}

			for _, secReq := range *secReqsP {
				// Authorino AuthConfig currently only supports one identity method for each identity evaluator.
				// It does not support, for instance, auth based on two api keys or api key AND oidc.
				// Thus, some OpenAPI 3.X security requirements are not supported:
				//
				// Not Supported:
				// security:
				//   - petstore_api_key: []
				//     toystore_api_key: []
				//     toystore_oidc: []
				//
				// Supported:
				// security:
				//   - petstore_api_key: []
				//   - toystore_api_key: []
				//   - toystore_oidc: []
				//

				// scopes not being used now
				for secSchemeName := range secReq {

					secSchemeI, err := oasDoc.Components.SecuritySchemes.JSONLookup(secSchemeName)
					if err != nil {
						return nil, err
					}

					secScheme := secSchemeI.(*openapi3.SecurityScheme) // panic if assertion fails

					identity, err := AuthConfigIdentityFromSecurityRequirement(
						operation.OperationID, // TODO(eastizle): OperationID can be null, fallback to some custom name
						path, opVerb, workloadName, secScheme)
					if err != nil {
						return nil, err
					}

					identities = append(identities, identity)
					// currently only support for one schema per requirement
					break
				}
			}

		}
	}
	return identities, nil
}

func AuthConfigIdentityFromSecurityRequirement(name, opPath, opVerb, workloadName string, secScheme *openapi3.SecurityScheme) (*authorinov1beta1.Identity, error) {
	if secScheme == nil {
		return nil, fmt.Errorf("sec scheme nil for operation path:%s method:%s", opPath, opVerb)
	}

	identity := &authorinov1beta1.Identity{
		Name:       name,
		Conditions: AuthConfigConditionsFromOperation(opPath, opVerb),
	}

	switch secScheme.Type {
	case "apiKey":
		AuthConfigIdentityFromApiKeyScheme(identity, secScheme, workloadName)
	case "openIdConnect":
		AuthConfigIdentityFromOIDCScheme(identity, secScheme)
	default:
		return nil, fmt.Errorf("sec scheme type %s not supported for path:%s method:%s", secScheme.Type, opPath, opVerb)
	}

	return identity, nil
}

func AuthConfigConditionsFromOperation(opPath, opVerb string) []authorinov1beta1.JSONPattern {
	return []authorinov1beta1.JSONPattern{
		{
			JSONPatternExpression: authorinov1beta1.JSONPatternExpression{
				Selector: "context.request.http.path",
				Operator: "eq",
				Value:    opPath,
			},
		},
		{
			JSONPatternExpression: authorinov1beta1.JSONPatternExpression{
				Selector: "context.request.http.method.@case:lower",
				Operator: "eq",
				Value:    strings.ToLower(opVerb),
			},
		},
	}
}

func AuthConfigIdentityFromApiKeyScheme(identity *authorinov1beta1.Identity, secScheme *openapi3.SecurityScheme, workloadName string) {
	// Fixed label selector for now
	apikey := authorinov1beta1.Identity_APIKey{
		LabelSelectors: map[string]string{
			"authorino.kuadrant.io/managed-by": "authorino",
			"app":                              workloadName,
		},
	}

	// TODO(eastizle): missing "cookie"
	switch secScheme.In {
	case "header":
		identity.Credentials.In = authorinov1beta1.Credentials_In("custom_header")
	case "query":
		identity.Credentials.In = authorinov1beta1.Credentials_In("query")
	}

	identity.Credentials.KeySelector = secScheme.Name
	identity.APIKey = &apikey
}

func AuthConfigIdentityFromOIDCScheme(identity *authorinov1beta1.Identity, secScheme *openapi3.SecurityScheme) {
	identity.Oidc = &authorinov1beta1.Identity_OidcConfig{
		Endpoint: secScheme.OpenIdConnectUrl,
	}
}

func AuthConfigResponsesFromOpenAPI(oasDoc *openapi3.T) ([]*authorinov1beta1.Response, error) {
	responses := make([]*authorinov1beta1.Response, 0)

	for path, pathItem := range oasDoc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			secReqsP := utils.OpenAPIOperationSecRequirements(oasDoc, operation)
			for _, secReq := range *secReqsP {
				// Authorino AuthConfig currently only supports one identity method for each identity evaluator.
				// It does not support, for instance, auth based on two api keys or api key AND oidc.
				// Thus, some OpenAPI 3.X security requirements are not supported:
				//
				// Not Supported:
				// security:
				//   - petstore_api_key: []
				//     toystore_api_key: []
				//     toystore_oidc: []
				//
				// Supported:
				// security:
				//   - petstore_api_key: []
				//   - toystore_api_key: []
				//   - toystore_oidc: []
				//

				// scopes not being used now
				for secSchemeName := range secReq {
					secSchemeI, err := oasDoc.Components.SecuritySchemes.JSONLookup(secSchemeName)
					if err != nil {
						return nil, err
					}

					secScheme := secSchemeI.(*openapi3.SecurityScheme) // panic if assertion fails

					if secScheme == nil {
						return nil, fmt.Errorf("sec scheme nil for operation path:%s method:%s", path, opVerb)
					}

					response := &authorinov1beta1.Response{
						Wrapper:    authorinov1beta1.Response_Wrapper("envoyDynamicMetadata"),
						WrapperKey: "ext_auth_data",
						Conditions: AuthConfigConditionsFromOperation(path, opVerb),
					}

					switch secScheme.Type {
					case "apiKey":
						response.Name = "rate-limit-apikey"
						response.JSON = &authorinov1beta1.Response_DynamicJSON{
							Properties: []authorinov1beta1.JsonProperty{
								{
									Name: "user-id",
									ValueFrom: authorinov1beta1.ValueFromAuthJSON{
										AuthJSON: `auth.identity.metadata.annotations.secret\.kuadrant\.io/user-id`,
									},
								},
							},
						}
					case "openIdConnect":
						response.Name = "rate-limit-oidc"
						response.JSON = &authorinov1beta1.Response_DynamicJSON{
							Properties: []authorinov1beta1.JsonProperty{
								{
									Name: "user-id",
									ValueFrom: authorinov1beta1.ValueFromAuthJSON{
										AuthJSON: `auth.identity.sub`,
									},
								},
							},
						}
					default:
						return nil, fmt.Errorf("sec scheme type %s not supported for path:%s method:%s", secScheme.Type, path, opVerb)
					}

					responses = append(responses, response)
					// currently only support for one schema per requirement
					break
				}
			}
		}
	}

	return responses, nil
}

func AuthConfigAPIKeyResponse() *authorinov1beta1.Response {
	return &authorinov1beta1.Response{
		Name:       "rate-limit-apikey",
		Wrapper:    authorinov1beta1.Response_Wrapper("envoyDynamicMetadata"),
		WrapperKey: "ext_auth_data",
		JSON: &authorinov1beta1.Response_DynamicJSON{
			Properties: []authorinov1beta1.JsonProperty{
				{
					Name: "user-id",
					ValueFrom: authorinov1beta1.ValueFromAuthJSON{
						AuthJSON: `auth.identity.metadata.annotations.secret\.kuadrant\.io/user-id`,
					},
				},
			},
		},
	}
}

func APIKeySecrets(workloadName string, api *gigglev1alpha1.Api) []*v1.Secret {
	secrets := make([]*v1.Secret, 0)
	for userID := range api.Spec.Users {
		if api.Spec.Users[userID].APIKey != nil {
			secrets = append(secrets, APIKeySecretsFromUser(userID, workloadName, *api.Spec.Users[userID].APIKey))
		}
	}
	return secrets
}

func APIKeySecretsFromUser(userID, workloadName, apiKey string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-apikey", workloadName, userID),
			Labels: map[string]string{
				"authorino.kuadrant.io/managed-by": "authorino",
				"app":                              workloadName,
			},
			Annotations: map[string]string{
				"secret.kuadrant.io/user-id": userID,
			},
		},
		StringData: map[string]string{
			"api_key": apiKey,
		},
		Type: v1.SecretTypeOpaque,
	}
}
