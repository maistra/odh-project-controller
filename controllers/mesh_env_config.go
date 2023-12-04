package controllers

import (
	"os"
)

const (
	MeshNamespaceEnv = "MESH_NAMESPACE"
	ControlPlaneEnv  = "CONTROL_PLANE_NAME"
)

func getControlPlaneName() string {
	return getEnvOr(ControlPlaneEnv, "basic")
}

func getMeshNamespace() string {
	return getEnvOr(MeshNamespaceEnv, "istio-system")
}

func getEnvOr(key, defaultValue string) string {
	if env, defined := os.LookupEnv(key); defined {
		return env
	}

	return defaultValue
}
