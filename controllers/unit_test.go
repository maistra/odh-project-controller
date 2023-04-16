package controllers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opendatahub-io/odh-project-controller/controllers"
	"github.com/opendatahub-io/odh-project-controller/test/labels"
)

var _ = When("Preparing host URL for Authorino's AuthConfig", Label(labels.Unit), func() {

	DescribeTable("it should remove protocol prefix from provided string",
		func(value, expected string) {
			Expect(controllers.RemoveProtocolPrefix(value)).To(Equal(expected))
		},
		Entry("for HTTP url", "http://authconfig.dev", "authconfig.dev"),
		Entry("for HTTPS url", "http://authconfig.dev", "authconfig.dev"),
		Entry("for HTTPS url", "authconfig.dev", "authconfig.dev"),
		Entry("for HTTPS url", "gopher://authconfig.dev", "gopher://authconfig.dev"),
	)

})
