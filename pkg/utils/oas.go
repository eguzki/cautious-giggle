package utils

import "github.com/getkin/kin-openapi/openapi3"

func OpenAPIOperationSecRequirements(oasDoc *openapi3.T, operation *openapi3.Operation) *openapi3.SecurityRequirements {
	if operation.Security == nil {
		return &oasDoc.Security
	}
	return operation.Security
}
