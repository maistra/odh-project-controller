package controllers

import (
	"context"
	"reflect"
	"strconv"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	maistrav1 "maistra.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconcile will manage the creation, update and deletion of the MeshMember for created the namespace.
func (r *OpenshiftServiceMeshReconciler) reconcileMeshMember(ctx context.Context, namespace *v1.Namespace) error {
	log := r.Log.WithValues("feature", "mesh", "namespace", namespace.Name)

	desiredMeshMember := newServiceMeshMember(namespace)
	foundMember := &maistrav1.ServiceMeshMember{}
	justCreated := false

	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredMeshMember.Name,
		Namespace: namespace.Name,
	}, foundMember)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Adding namespace to the mesh")

			err = r.Create(ctx, desiredMeshMember)
			if err != nil && !apierrs.IsAlreadyExists(err) {
				log.Error(err, "Unable to create ServiceMeshMember")

				return errors.Wrap(err, "unable to create ServiceMeshMember")
			}

			justCreated = true
		} else {
			log.Error(err, "Unable to fetch the ServiceMeshMember")

			return errors.Wrap(err, "unable to fetch the ServiceMeshMember")
		}
	}

	// Reconcile the membership spec if it has been manually modified
	if !justCreated && !compareMeshMembers(*desiredMeshMember, *foundMember) {
		log.Info("Reconciling ServiceMeshMember")

		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Get(ctx, types.NamespacedName{
				Name:      desiredMeshMember.Name,
				Namespace: namespace.Name,
			}, foundMember); err != nil {
				return errors.Wrapf(err, "failed getting ServieMeshMember %s in namespace %s", desiredMeshMember.Name, namespace.Name)
			}

			foundMember.Spec = desiredMeshMember.Spec
			foundMember.ObjectMeta.Labels = desiredMeshMember.ObjectMeta.Labels

			return errors.Wrap(r.Update(ctx, foundMember), "failed updating ServiceMeshMember")
		})

		if err != nil {
			log.Error(err, "Unable to reconcile the ServiceMeshMember")

			return errors.Wrap(err, "unable to reconcile the ServiceMeshMember")
		}
	}

	return nil
}

func newServiceMeshMember(namespace *v1.Namespace) *maistrav1.ServiceMeshMember {
	controlPlaneName := getControlPlaneName()
	meshNamespace := getMeshNamespace()

	return &maistrav1.ServiceMeshMember{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default", // The name MUST be default, per the maistra docs
			Namespace: namespace.Name,
		},
		Spec: maistrav1.ServiceMeshMemberSpec{
			ControlPlaneRef: maistrav1.ServiceMeshControlPlaneRef{
				Name:      controlPlaneName,
				Namespace: meshNamespace,
			},
		},
	}
}

func compareMeshMembers(m1, m2 maistrav1.ServiceMeshMember) bool {
	return reflect.DeepEqual(m1.ObjectMeta.Labels, m2.ObjectMeta.Labels) &&
		reflect.DeepEqual(m1.Spec, m2.Spec)
}

func serviceMeshIsNotEnabled(meta metav1.ObjectMeta) bool {
	serviceMeshAnnotation := meta.Annotations[AnnotationServiceMesh]
	if serviceMeshAnnotation != "" {
		enabled, _ := strconv.ParseBool(serviceMeshAnnotation)

		return !enabled
	}

	return true
}

func (r *OpenshiftServiceMeshReconciler) findIstioIngress(ctx context.Context) (routev1.RouteList, error) {
	meshNamespace := getMeshNamespace()

	routes := routev1.RouteList{}
	if err := r.List(ctx, &routes, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{"app": "odh-dashboard"}),
		Namespace:     meshNamespace,
	}); err != nil {
		r.Log.Error(err, "Unable to find matching gateway")

		return routev1.RouteList{}, errors.Wrap(err, "unable to find matching gateway")
	}

	if len(routes.Items) == 0 {
		route := &routev1.Route{}

		return routes, apierrs.NewNotFound(schema.GroupResource{
			Group:    route.GroupVersionKind().Group,
			Resource: route.ResourceVersion,
		}, "no-route-matching-label")
	}

	return routes, nil
}
