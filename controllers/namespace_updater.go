package controllers

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func (r *OpenshiftServiceMeshReconciler) addGatewayAnnotations(ctx context.Context, ns *v1.Namespace) error {
	if ns.ObjectMeta.Annotations[AnnotationGatewayHost] != "" && ns.ObjectMeta.Annotations[AnnotationGateway] != "" {
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

	ns.ObjectMeta.Annotations[AnnotationGatewayHost] = ExtractHostName(routes.Items[0].Spec.Host)
	gateway := extractGateway(routes.Items[0].ObjectMeta)
	if gateway != "" {
		ns.ObjectMeta.Annotations[AnnotationGateway] = gateway
	}
	return r.Client.Update(ctx, ns)
}

func extractGateway(meta metav1.ObjectMeta) string {
	gwName := meta.Labels[LabelMaistraGw]
	if gwName == "" {
		return ""
	}
	gateway := gwName

	gwNs := meta.Labels[LabelMaistraGwNs]
	if gwNs != "" {
		gateway = gwNs + "/" + gwName
	}

	return gateway
}
