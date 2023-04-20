package controllers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opendatahub-io/odh-project-controller/controllers"
	"github.com/opendatahub-io/odh-project-controller/test/labels"
)

var _ = Describe("Controller helper functions", Label(labels.Unit), func() {

	When("Preparing host URL for Authorino's AuthConfig", func() {

		DescribeTable("it should remove protocol prefix from provided string and path",
			func(value, expected string) {
				Expect(controllers.ExtractHostName(value)).To(Equal(expected))
			},
			Entry("for HTTP url", "http://authconfig.dev", "authconfig.dev"),
			Entry("for HTTP url with path", "http://authconfig.dev/api/resources", "authconfig.dev"),
			Entry("for HTTPS url", "http://authconfig.dev", "authconfig.dev"),
			Entry("for HTTPS url with path and query params", "http://authconfig.dev/api/resources?limit=500", "authconfig.dev"),
			Entry("for HTTPS url", "authconfig.dev", "authconfig.dev"),
			Entry("for HTTPS url", "gopher://authconfig.dev", "gopher://authconfig.dev"),
		)

	})

	When("Checking namespace", func() {
		DescribeTable("it should not process reserved namespaces",
			func(ns string, expected bool) {
				Expect(controllers.IsReservedNamespace(ns)).To(Equal(expected))
			},
			Entry("kube-system is reserved namespace", "kube-system", true),
			Entry("openshift-build is reserved namespace", "openshift-build", true),
			Entry("kube-public is reserved namespace", "kube-public", true),
			Entry("openshift-infra is reserved namespace", "openshift-infra", true),
			Entry("kube-node-lease is reserved namespace", "kube-node-lease", true),
			Entry("openshift is reserved namespace", "openshift", true),
			Entry("istio-system is reserved namespace", "istio-system", true),
			Entry("openshift-authentication is reserved namespace", "openshift-authentication", true),
			Entry("openshift-apiserver is reserved namespace", "openshift-apiserver", true),
			Entry("mynamespace is not reserved namespace", "mynamespace", false),
			Entry("openshiftmynamespace is not reserved namespace", "openshiftmynamespace", false),
			Entry("kubemynamespace is not reserved namespace", "kubemynamespace", false),
			Entry("istio-system-openshift is not reserved namespace", "istio-system-openshift ", false),
		)
	})

})
