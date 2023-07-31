package controllers_test

import (
	"context"
	"os"
	"time"

	authorino "github.com/kuadrant/authorino/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opendatahub-io/odh-project-controller/controllers"
	"github.com/opendatahub-io/odh-project-controller/test"
	. "github.com/opendatahub-io/odh-project-controller/test/cluster"
	"github.com/opendatahub-io/odh-project-controller/test/labels"
	openshiftv1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
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
		istioNs,
		testNs *corev1.Namespace
		objectCleaner *Cleaner
		route         *openshiftv1.Route
	)

	BeforeEach(func() {
		objectCleaner = CreateCleaner(cli, envTest.Config, timeout, interval)
		istioNs = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "istio-system",
			},
		}

		weight := int32(100)
		route = &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "odh-gateway",
				Namespace: "istio-system",
				Labels: map[string]string{
					"app":                                    "odh-dashboard",
					controllers.LabelMaistraGatewayName:      "odh-gateway",
					controllers.LabelMaistraGatewayNamespace: "opendatahub",
				},
			},
			Spec: openshiftv1.RouteSpec{
				Host: "istio.io",
				To: openshiftv1.RouteTargetReference{
					Name:   "istio-ingressgateway",
					Weight: &weight,
				},
			},
		}

		Expect(cli.Create(context.Background(), istioNs)).To(Succeed())
		Expect(cli.Create(context.Background(), route)).To(Succeed())
	})

	AfterEach(func() {
		objectCleaner.DeleteAll(istioNs, route, testNs)
	})

	Context("enabling service mesh", func() {

		It("should register it in the mesh if annotation is set to true", func() {
			// given
			testNs = &corev1.Namespace{
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
			testNs = &corev1.Namespace{
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

		It("should create an SMM with default name and ns when no env vars defined", func() {
			// given
			testNs = &corev1.Namespace{
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
				Expect(member.Spec.ControlPlaneRef.Name).To(Equal("basic"))
				Expect(member.Spec.ControlPlaneRef.Namespace).To(Equal("istio-system"))
			})
		})

		It("should create an SMM with specified name defined in env var", func() {
			// given
			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "meshified-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			_ = os.Setenv(controllers.ControlPlaneEnv, "minimal")
			defer os.Unsetenv(controllers.ControlPlaneEnv)
			_ = os.Setenv(controllers.MeshNamespaceEnv, "istio-system")
			defer os.Unsetenv(controllers.MeshNamespaceEnv)

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
				Expect(member.Spec.ControlPlaneRef.Name).To(Equal("minimal"))
				Expect(member.Spec.ControlPlaneRef.Namespace).To(Equal("istio-system"))
			})
		})
	})

	Context("enabling external authorization", func() {

		It("should configure authorization using defaults for ns belonging to the mesh", func() {
			// given
			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "meshified-and-authorized-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			defer objectCleaner.DeleteAll(testNs)

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			By("creating authorization config resource", func() {
				expectedAuthConfig := &authorino.AuthConfig{}

				Expect(controllers.ConvertToStructuredResource(test.ExpectedAuthConfig, expectedAuthConfig)).To(Succeed())
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

				Expect(actualAuthConfig.Labels).To(Equal(expectedAuthConfig.Labels))
				Expect(actualAuthConfig.Name).To(Equal(testNs.GetName() + "-protection"))
				Expect(actualAuthConfig.Spec.Hosts).To(Equal(expectedAuthConfig.Spec.Hosts))
			})
		})

		It("should configure authorization using env vars for ns belonging to the mesh", func() {
			// given
			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "meshified-and-authorized-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			defer objectCleaner.DeleteAll(testNs)

			// when
			_ = os.Setenv(controllers.AuthorinoLabelSelector, "app=rhods")
			defer os.Unsetenv(controllers.AuthorinoLabelSelector)
			_ = os.Setenv(controllers.AuthAudience, "opendatahub.io,foo , bar")
			defer os.Unsetenv(controllers.AuthAudience)

			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			By("creating authorization config resource", func() {
				expectedAuthConfig := &authorino.AuthConfig{}

				Expect(controllers.ConvertToStructuredResource(test.ExpectedAuthConfig, expectedAuthConfig)).To(Succeed())
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

				Expect(actualAuthConfig.Name).To(Equal(testNs.GetName() + "-protection"))
				Expect(actualAuthConfig.Labels).To(HaveKeyWithValue("app", "rhods"))
				Expect(actualAuthConfig.Spec.Hosts).To(Equal(expectedAuthConfig.Spec.Hosts))
				Expect(actualAuthConfig.Spec.Identity[0].KubernetesAuth.Audiences).
					To(And(HaveLen(3), ContainElements("opendatahub.io", "foo", "bar")))
			})
		})

		It("should not configure authorization rules if namespace is not part of the mesh", func() {
			// given
			testNs = &corev1.Namespace{
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

	Context("propagating service mesh gateway info", func() {

		It("should add just gateway name to the namespace if there is no gateway namespace defined", func() {
			// given
			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "plain-meshified-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			// update route to remove gateway namespace label
			delete(route.Labels, controllers.LabelMaistraGatewayNamespace)
			Expect(cli.Update(context.Background(), route)).To(Succeed())

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			actualTestNs := &corev1.Namespace{}
			Eventually(func() string {
				_ = cli.Get(context.Background(), types.NamespacedName{Name: testNs.Name}, actualTestNs)

				return actualTestNs.Annotations[controllers.AnnotationPublicGatewayName]
			}).
				WithTimeout(timeout).
				WithPolling(interval).
				Should(Equal("odh-gateway"))
		})

		It("should add fully qualified gateway name to the namespace", func() {
			// given
			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "plain-meshified-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			actualTestNs := &corev1.Namespace{}
			Eventually(func() string {
				_ = cli.Get(context.Background(), types.NamespacedName{Name: testNs.Name}, actualTestNs)

				return actualTestNs.Annotations[controllers.AnnotationPublicGatewayName]
			}).
				WithTimeout(timeout).
				WithPolling(interval).
				Should(Equal("opendatahub/odh-gateway"))
		})

		It("should add external gateway host to the namespace", func() {
			// given
			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "plain-meshified-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			actualTestNs := &corev1.Namespace{}
			Eventually(func() string {
				_ = cli.Get(context.Background(), types.NamespacedName{Name: testNs.Name}, actualTestNs)

				return actualTestNs.Annotations[controllers.AnnotationPublicGatewayExternalHost]
			}).
				WithTimeout(timeout).
				WithPolling(interval).
				Should(Equal("istio.io"))
		})

		It("should add internal gateway host to the namespace", func() {
			// given
			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "plain-meshified-ns",
					Annotations: map[string]string{
						controllers.AnnotationServiceMesh: "true",
					},
				},
			}

			// when
			Expect(cli.Create(context.Background(), testNs)).To(Succeed())

			// then
			actualTestNs := &corev1.Namespace{}
			Eventually(func() string {
				_ = cli.Get(context.Background(), types.NamespacedName{Name: testNs.Name}, actualTestNs)

				return actualTestNs.Annotations[controllers.AnnotationPublicGatewayInternalHost]
			}).
				WithTimeout(timeout).
				WithPolling(interval).
				Should(Equal("istio-ingressgateway.istio-system.svc.cluster.local"))
		})

	})

})
