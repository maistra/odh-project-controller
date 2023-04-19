package controllers_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/onsi/gomega/format"
	. "github.com/opendatahub-io/odh-project-controller/test/cluster"
	"github.com/opendatahub-io/odh-project-controller/test/labels"

	authorino "github.com/kuadrant/authorino/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opendatahub-io/odh-project-controller/controllers"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	maistrav1 "maistra.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeout  = 1 * time.Minute
	interval = 250 * time.Millisecond
)

var _ = When("Namespace is created", Label(labels.EvnTest), func() {

	var (
		testNs *v1.Namespace
	)

	AfterEach(func() {
		CreateCleaner(cli, envTest.Config, timeout, interval).DeleteAll(testNs)
	})

	Context("enabling service mesh", func() {

		It("should register it in the mesh if annotation is set to true", func() {
			// given
			testNs = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "meshified-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			By("creating service mesh member object in the namespace", func() {
				member := &maistrav1.ServiceMeshMember{}
				namespacedName := types.NamespacedName{
					Namespace: testNs.Name,
					Name:      "default",
				}
				Eventually(func() error {
					return cli.Get(context.Background(), namespacedName, member)
				}).
					WithTimeout(timeout).
					WithPolling(interval).
					Should(Succeed())
			})
		})

		It("should not register it in the mesh if annotation is absent", func() {
			// given
			testNs = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-meshified-namespace",
				},
			}

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			By("ensuring no service mesh member created", func() {
				members := &maistrav1.ServiceMeshMemberList{}

				Eventually(func() error {
					return cli.List(context.Background(), members, client.InNamespace(testNs.Name))
				}).
					WithTimeout(timeout).
					WithPolling(interval).
					Should(Succeed())

				Expect(members.Items).Should(BeEmpty())
			})
		})
	})

	Context("enabling external authorization", func() {

		XIt("should configure authorization rules for ns belonging to the mesh", func() {
			// given
			testNs = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "meshified-and-authorized-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			By("creating authorization config resource", func() {
				expectedAuthConfig := &authorino.AuthConfig{}
				file, _ := os.ReadFile("testdata/expected_authconfig.yaml")
				Expect(controllers.ConvertToStructuredResource(file, expectedAuthConfig)).To(Succeed())
				namespacedName := types.NamespacedName{
					Namespace: testNs.Name,
					Name:      expectedAuthConfig.Name,
				}
				actualAuthConfig := &authorino.AuthConfig{}
				Eventually(func() error {
					return cli.Get(context.Background(), namespacedName, actualAuthConfig)
				}).
					WithTimeout(timeout).
					WithPolling(interval).
					Should(Succeed())

				Expect(controllers.CompareAuthConfigs(*expectedAuthConfig, *actualAuthConfig)).
					To(BeTrue(), fmt.Sprintf("Expected: %1s\n Got: %2s\n", format.Object(expectedAuthConfig, 1), format.Object(actualAuthConfig, 1)))
			})
		})

		It("should not configure authorization rules if namespace is not part of the mesh", func() {
			// given
			testNs = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-meshified-nor-authorized-namespace",
				},
			}

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			By("ensuring no authorization config has been created", func() {
				authConfigs := &authorino.AuthConfigList{}

				Eventually(func() error {
					return cli.List(context.Background(), authConfigs, client.InNamespace(testNs.Name))
				}).
					WithTimeout(timeout).
					WithPolling(interval).
					Should(Succeed())

				Expect(authConfigs.Items).Should(BeEmpty())
			})
		})
	})

})
