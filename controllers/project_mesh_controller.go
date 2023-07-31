package controllers

import (
	"context"
	"regexp"

	"github.com/go-logr/logr"
	"github.com/kuadrant/authorino/api/v1beta1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8serrs "k8s.io/apimachinery/pkg/util/errors"
	maistrav1 "maistra.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OpenshiftServiceMeshReconciler holds the controller configuration.
type OpenshiftServiceMeshReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=authorino.kuadrant.io,resources=authconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=maistra.io,resources=servicemeshmembers;servicemeshmembers/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=maistra.io,resources=servicemeshcontrolplanes,verbs=get;list;watch;create;update;patch;use
// +kubebuilder:rbac:groups=maistra.io,resources=servicemeshcontrolplanes,verbs=get;list;watch;create;update;patch;use
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch

// Reconcile ensures that the namespace has all required resources needed to be part of the Service Mesh of Open Data Hub.
func (r *OpenshiftServiceMeshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)

	reconcilers := []reconcileFunc{r.addGatewayAnnotations, r.reconcileMeshMember, r.reconcileAuthConfig}

	namespace := &v1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, namespace); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Stopping reconciliation")

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrap(err, "failed getting namespace")
	}

	if IsReservedNamespace(namespace.Name) || serviceMeshIsNotEnabled(namespace.ObjectMeta) {
		log.Info("Skipped")

		return ctrl.Result{}, nil
	}

	var errs []error
	for _, reconciler := range reconcilers {
		errs = append(errs, reconciler(ctx, namespace))
	}

	return ctrl.Result{}, k8serrs.NewAggregate(errs)
}

func (r *OpenshiftServiceMeshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck //reason there is no point in wrapping it
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Namespace{}).
		Owns(&maistrav1.ServiceMeshMember{}).
		Owns(&v1beta1.AuthConfig{}).
		Complete(r)
}

var reservedNamespaceRegex = regexp.MustCompile(`^(openshift|istio-system)$|^(kube|openshift)-.*$`)

func IsReservedNamespace(namepace string) bool {
	return reservedNamespaceRegex.MatchString(namepace)
}
