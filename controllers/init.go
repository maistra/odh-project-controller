package controllers

import (
	authorino "github.com/kuadrant/authorino/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	maistrav1 "maistra.io/api/core/v1"
)

// RegisterSchemes adds schemes of used resources to controller's scheme.
func RegisterSchemes(s *runtime.Scheme) {
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(maistrav1.AddToScheme(s))
	utilruntime.Must(authorino.AddToScheme(s))
}
