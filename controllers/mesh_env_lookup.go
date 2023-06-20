package controllers

import (
	"os"
)

const (
	MeshNamespaceEnv = "MESH_NAMESPACE"
	ControlPlaneEnv  = "CONTROL_PLANE_NAME"
)

// Helper functions to fetch the relevant environment variables
func getControlPlaneName() string {
	controlPlaneName := "basic"
	if env, defined := os.LookupEnv(ControlPlaneEnv); defined {
		controlPlaneName = env
	}
	return controlPlaneName
}

func getMeshNamespace() string {
	meshNamespace := "istio-system"
	if env, defined := os.LookupEnv(MeshNamespaceEnv); defined {
		meshNamespace = env
	}
	return meshNamespace
}
