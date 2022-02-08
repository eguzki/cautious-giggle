package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"k8s.io/apimachinery/pkg/util/validation"
)

var (
	// NonAlphanumRegexp not alphanumeric
	NonAlphanumRegexp = regexp.MustCompile(`[^0-9A-Za-z]`)
)

func OpenAPIOperationSecRequirements(oasDoc *openapi3.T, operation *openapi3.Operation) *openapi3.SecurityRequirements {
	if operation.Security == nil {
		return &oasDoc.Security
	}
	return operation.Security
}

func K8sNameFromOpenAPITitle(obj *openapi3.T) (string, error) {
	openapiTitle := obj.Info.Title
	openapiTitleToLower := strings.ToLower(openapiTitle)
	objName := NonAlphanumRegexp.ReplaceAllString(openapiTitleToLower, "")

	// DNS Subdomain Names
	// If the name would be part of some label, validation would be DNS Label Names (validation.IsDNS1123Label)
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
	errStrings := validation.IsDNS1123Subdomain(objName)
	if len(errStrings) > 0 {
		errStr := strings.Join(errStrings, ",")
		return "", fmt.Errorf("k8s name from OAS not valid: %s", errStr)
	}
	return objName, nil
}
