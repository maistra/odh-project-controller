package controllers

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	MeshNamespaceEnv       = "MESH_NAMESPACE"
	ControlPlaneEnv        = "CONTROL_PLANE_NAME"
	AuthorinoLabelSelector = "AUTHORINO_LABEL"
)

func getControlPlaneName() string {
	return getEnvOr(ControlPlaneEnv, "basic")
}

func getMeshNamespace() string {
	return getEnvOr(MeshNamespaceEnv, "istio-system")
}

func getAuthorinoTopic() ([]string, error) {
	topic := getEnvOr(AuthorinoLabelSelector, "authorino/topic=odh")
	keyValue := strings.Split(topic, "=")
	if len(keyValue) != 2 {
		return nil, errors.Errorf("Expected authorino topic to be in key=value format, got [%s]", topic)
	}
	return keyValue, nil
}

func getEnvOr(key, defaultValue string) string {
	if env, defined := os.LookupEnv(key); defined {
		return env
	}

	return defaultValue
}
