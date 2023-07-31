package controllers

const (
	AnnotationServiceMesh                            = "opendatahub.io/service-mesh"
	AnnotationPublicGatewayName                      = "service-mesh.opendatahub.io/public-gateway-name"
	AnnotationPublicGatewayExternalHost              = "service-mesh.opendatahub.io/public-gateway-host-external"
	AnnotationPublicGatewayInternalHost              = "service-mesh.opendatahub.io/public-gateway-host-internal"
	AnnotationProjectModelGatewayHostPatternExternal = "service-mesh.opendatahub.io/model-gateway-hostpattern-external"
	AnnotationProjectModelGatewayHostPatternInternal = "service-mesh.opendatahub.io/model-gateway-hostpattern-internal"
	LabelMaistraGatewayName                          = "maistra.io/gateway-name"
	LabelMaistraGatewayNamespace                     = "maistra.io/gateway-namespace"
)
