package controllers

import (
	projectv1 "github.com/openshift/api/project/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	maistrav1 "maistra.io/api/core/v1"
)

func RegisterSchemes(s *runtime.Scheme) {
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(maistrav1.AddToScheme(s))
	utilruntime.Must(projectv1.AddToScheme(s))
}
