package controllers

import (
	"context"
	"reflect"
	"strconv"

	v1 "k8s.io/api/core/v1"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	maistrav1 "maistra.io/api/core/v1"
)

// Reconcile will manage the creation, update and deletion of the MeshMember for created the namespace.
func (r *OpenshiftServiceMeshReconciler) reconcileMeshMember(ctx context.Context, ns *v1.Namespace) error {
	log := r.Log.WithValues("feature", "mesh", "namespace", ns.Name)

	if IsReservedNamespace(ns.Name) || serviceMeshIsNotEnabled(ns.ObjectMeta) {
		log.Info("Skipped")
		return nil
	}

	desiredMeshMember := newServiceMeshMember(ns)
	foundMember := &maistrav1.ServiceMeshMember{}
	justCreated := false
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredMeshMember.Name,
		Namespace: ns.Name,
	}, foundMember)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Adding namespace to the mesh")
			err = r.Create(ctx, desiredMeshMember)
			if err != nil && !apierrs.IsAlreadyExists(err) {
				log.Error(err, "Unable to create ServiceMeshMember")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "Unable to fetch the ServiceMeshMember")
			return err
		}
	}

	// Reconcile the membership spec if it has been manually modified
	if !justCreated && !compareMeshMembers(*desiredMeshMember, *foundMember) {
		log.Info("Reconciling ServiceMeshMember")
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Get(ctx, types.NamespacedName{
				Name:      desiredMeshMember.Name,
				Namespace: ns.Namespace,
			}, foundMember); err != nil {
				return err
			}
			foundMember.Spec = desiredMeshMember.Spec
			foundMember.ObjectMeta.Labels = desiredMeshMember.ObjectMeta.Labels
			return r.Update(ctx, foundMember)
		})
		if err != nil {
			log.Error(err, "Unable to reconcile the ServiceMeshMember")
			return err
		}
	}

	return nil
}

func newServiceMeshMember(ns *v1.Namespace) *maistrav1.ServiceMeshMember {
	return &maistrav1.ServiceMeshMember{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default", // The name MUST be default, per the maistra docs
			Namespace: ns.Name,
			Labels:    map[string]string{"opendatahub.io/ns": ns.Name},
		},
		Spec: maistrav1.ServiceMeshMemberSpec{
			ControlPlaneRef: maistrav1.ServiceMeshControlPlaneRef{
				Name:      "odh",
				Namespace: "istio-system",
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
