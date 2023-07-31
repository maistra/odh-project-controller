package controllers

import (
	authorino "github.com/kuadrant/authorino/api/v1beta1"
	configv1 "github.com/openshift/api/config/v1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	maistrav1 "maistra.io/api/core/v1"
)

// RegisterSchemes adds schemes of used resources to controller's scheme.
func RegisterSchemes(s *runtime.Scheme) {
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(maistrav1.AddToScheme(s))
	utilruntime.Must(routev1.Install(s))
	utilruntime.Must(configv1.Install(s))
	utilruntime.Must(authorino.AddToScheme(s))
}
