package istiogenerators

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	istionetworkingapi "istio.io/api/networking/v1beta1"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/eguzki/cautious-giggle/pkg/utils"
)

func VirtualService(doc *openapi3.T, serviceName, serviceNamespace string,
	servicePort uint32, gateways []string, publicHost string,
	pathMatchType string) (*istionetworking.VirtualService, error) {

	objectName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		return nil, err
	}

	destination := &istionetworkingapi.Destination{
		Host: fmt.Sprintf("%s.%s.svc", serviceName, serviceNamespace),
		Port: &istionetworkingapi.PortSelector{Number: servicePort},
	}

	vs := &istionetworking.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			// Missing namespace
			Name: objectName,
		},
		Spec: istionetworkingapi.VirtualService{
			Gateways: gateways,
			Hosts:    []string{publicHost},
			Http:     HTTPRoutesFromOpenAPI(doc, destination, pathMatchType),
		},
	}

	return vs, nil
}
