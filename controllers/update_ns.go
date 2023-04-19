package controllers

import (
	"context"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func (r *OpenshiftServiceMeshReconciler) addGatewayHostAnnotation(ctx context.Context, ns *v1.Namespace) error {
	if ns.ObjectMeta.Annotations[AnnotationGatewayHost] != "" {
		// If annotation is present we have nothing to do
		return nil
	}

	routes, err := r.findIstioIngress(ctx)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}
		// TODO rethink if we shouldn't just fail here
		r.Log.Info("Unable to find matching istio ingress gateway. Some things might not work.")
		return nil
	}

	ns.ObjectMeta.Annotations[AnnotationGatewayHost] = RemoveProtocolPrefix(routes.Items[0].Spec.Host)
	return r.Client.Update(ctx, ns)
}
