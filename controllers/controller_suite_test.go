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

package controllers // TODO this should be _test package

import (
	"bytes"
	"context"
	"github.com/manifestival/manifestival"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	maistramanifests "maistra.io/api/manifests"
	"path/filepath"
	"testing"
	"time"

	mf "github.com/manifestival/manifestival"

	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// +kubebuilder:docs-gen:collapse=Imports

var (
	cli     client.Client
	envTest *envtest.Environment
	ctx     context.Context
	cancel  context.CancelFunc
)

var testScheme = runtime.NewScheme()

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller & Webhook Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.TODO())

	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.TimeEncoderOfLayout(time.RFC3339),
	}
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseFlagOptions(&opts)))

	By("Bootstrapping k8s test environment")
	envTest = &envtest.Environment{
		CRDInstallOptions: envtest.CRDInstallOptions{
			CRDs:               loadCRDs(),
			Paths:              []string{filepath.Join("..", "config", "crd", "external")},
			ErrorIfPathMissing: true,
			CleanUpAfterUse:    false,
		},
	}

	cfg, err := envTest.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	RegisterSchemes(testScheme)

	//+kubebuilder:scaffold:scheme

	cli, err = client.New(cfg, client.Options{Scheme: testScheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(cli).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             testScheme,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	err = (&OpenshiftServiceMeshReconciler{
		Client: cli,
		Log:    ctrl.Log.WithName("controllers").WithName("project-controller"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed(), "Failed to start manager")
	}()
})

var _ = AfterSuite(func() {
	By("Tearing down the test environment")
	cancel()
	Expect(envTest.Stop()).To(Succeed())
})

func loadCRDs() []*v1.CustomResourceDefinition {
	smmYaml, err := maistramanifests.ReadManifest("maistra.io_servicemeshmembers.yaml")
	Expect(err).NotTo(HaveOccurred())
	crd := &v1.CustomResourceDefinition{}
	err = convertToStructuredResource(smmYaml, crd)
	Expect(err).NotTo(HaveOccurred())

	return []*v1.CustomResourceDefinition{crd}
}

func convertToStructuredResource(yamlContent []byte, out interface{}, opts ...manifestival.Option) error {
	reader := bytes.NewReader(yamlContent)
	m, err := mf.ManifestFrom(manifestival.Reader(reader), opts...)
	if err != nil {
		return err
	}

	err = scheme.Scheme.Convert(&m.Resources()[0], out, nil)
	if err != nil {
		return err
	}
	return nil
}
