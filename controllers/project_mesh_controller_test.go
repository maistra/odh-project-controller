/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	maistrav1 "maistra.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	duration = 1 * time.Minute
	interval = 250 * time.Millisecond
)

var _ = When("Namespace is created", func() {

	var testNs *v1.Namespace

	AfterEach(func() {
		Expect(cli.Delete(context.Background(), testNs)).To(Succeed())
	})

	It("should register it in the mesh if annotation is present", func() {
		// given
		testNs = &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "meshified-ns",
				Annotations: map[string]string{
					AnnotationServiceMesh: "true",
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
				WithTimeout(duration).
				WithPolling(interval).
				Should(Succeed())
		})
	})

	It("should not register it in the mesh if annotation is not present", func() {
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
				WithTimeout(duration).
				WithPolling(interval).
				Should(Succeed())

			Expect(members.Items).Should(BeEmpty())
		})
	})

})
