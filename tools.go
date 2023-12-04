//go:build tools
// +build tools

package tools

// nolint
import (
	_ "embed"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "github.com/onsi/ginkgo/v2/ginkgo/generators"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.) to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/code-generator"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
