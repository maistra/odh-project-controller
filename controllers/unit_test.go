package controllers_test

import (
	"github.com/opendatahub-io/odh-project-controller/controllers"
	"github.com/opendatahub-io/odh-project-controller/test/labels"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller helper functions", Label(labels.Unit), func() {

	When("Extracting host URL", func() {

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

})
