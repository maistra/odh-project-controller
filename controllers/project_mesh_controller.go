package controllers

import (
	"context"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/errors"
	maistrav1 "maistra.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AnnotationServiceMesh = "opendatahub.io/service-mesh"
	AnnotationHubURL      = "opendatahub.io/hub-url"
)

// OpenshiftServiceMeshReconciler holds the controller configuration.
type OpenshiftServiceMeshReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=maistra.io,resources=servicemeshmembers;servicemeshmembers/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=maistra.io,resources=servicemeshcontrolplanes,verbs=get;list;watch;create;update;patch;use
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch

// Reconcile TODO yeah.
func (r *OpenshiftServiceMeshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)

	reconcilers := []reconcileFunc{r.reconcileMeshMember, r.reconcileAuthConfig}

	ns := &v1.Namespace{}
	err := r.Get(ctx, req.NamespacedName, ns)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Stopping reconciliation")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	errs := make([]error, len(reconcilers))
	for _, reconciler := range reconcilers {
		errs = append(errs, reconciler(ctx, ns))
	}

	return ctrl.Result{}, errors.NewAggregate(errs)
}

func (r *OpenshiftServiceMeshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Namespace{}).
		Owns(&maistrav1.ServiceMeshMember{}).
		Complete(r)
}
