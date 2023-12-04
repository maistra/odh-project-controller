package controllers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *OpenshiftServiceMeshReconciler) addGatewayAnnotations(ctx context.Context, namespace *v1.Namespace) error {
	if namespace.ObjectMeta.Annotations[AnnotationPublicGatewayExternalHost] != "" &&
		namespace.ObjectMeta.Annotations[AnnotationPublicGatewayInternalHost] != "" &&
		namespace.ObjectMeta.Annotations[AnnotationPublicGatewayName] != "" &&
		namespace.ObjectMeta.Annotations[AnnotationProjectModelGatewayHostPatternExternal] != "" &&
		namespace.ObjectMeta.Annotations[AnnotationProjectModelGatewayHostPatternInternal] != "" {
		// If annotation is present we have nothing to do
		return nil
	}

	routes, err := r.findIstioIngress(ctx)
	if err != nil {
		r.Log.Error(err, "Unable to find matching istio ingress gateway.")

		return err
	}

	appDomain, err := r.findAppDomain(ctx)
	if err != nil {
		r.Log.Error(err, "unable to find app domain")

		return err
	}

	namespace.ObjectMeta.Annotations[AnnotationPublicGatewayExternalHost] = ExtractHostName(routes.Items[0].Spec.Host)
	namespace.ObjectMeta.Annotations[AnnotationPublicGatewayInternalHost] = fmt.Sprintf("%s.%s.svc.cluster.local", routes.Items[0].Spec.To.Name, getMeshNamespace())

	namespace.ObjectMeta.Annotations[AnnotationProjectModelGatewayHostPatternExternal] = fmt.Sprintf("*.%s.%s", namespace.Name, appDomain)
	namespace.ObjectMeta.Annotations[AnnotationProjectModelGatewayHostPatternInternal] = fmt.Sprintf("*.%s.svc.cluster.local", namespace.Name)

	gateway := extractGateway(routes.Items[0].ObjectMeta)
	if gateway != "" {
		namespace.ObjectMeta.Annotations[AnnotationPublicGatewayName] = gateway
	}

	return errors.Wrap(r.Client.Update(ctx, namespace), "failed updating namespace with annotations")
}

func extractGateway(meta metav1.ObjectMeta) string {
	gwName := meta.Labels[LabelMaistraGatewayName]
	if gwName == "" {
		return ""
	}

	gateway := gwName

	gwNs := meta.Labels[LabelMaistraGatewayNamespace]
	if gwNs != "" {
		gateway = gwNs + "/" + gwName
	}

	return gateway
}
