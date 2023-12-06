package controllers_test

import (
	"github.com/opendatahub-io/odh-project-controller/controllers"
	"github.com/opendatahub-io/odh-project-controller/test/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller predicates functions", Label(labels.Unit), func() {

	When("Checking namespace", func() {

		It("should trigger update event when annotation has been removed", func() {
			// given
			meshAwareNamespaces := controllers.MeshAwareNamespaces()

			// when
			namespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-with-annotation-removed",
				},
			}
			namespaceWithAnnotation := corev1.Namespace{}
			namespace.DeepCopyInto(&namespaceWithAnnotation)
			namespaceWithAnnotation.SetAnnotations(map[string]string{
				controllers.AnnotationServiceMesh: "true",
			})

			annotationRemovedEvent := event.UpdateEvent{
				ObjectOld: &namespaceWithAnnotation,
				ObjectNew: &namespace,
			}

			// then
			Expect(meshAwareNamespaces.UpdateFunc(annotationRemovedEvent)).To(BeTrue())
		})

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
