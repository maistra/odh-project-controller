package controllers

import (
	"regexp"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func MeshAwareNamespaces() predicate.Funcs {
	filter := func(object client.Object) bool {
		return !IsReservedNamespace(object.GetName()) && object.GetAnnotations()[AnnotationServiceMesh] != ""
	}

	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return filter(createEvent.Object)
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			if annotationRemoved(updateEvent, AnnotationServiceMesh) {
				// annotation has been just removed, handle it in reconcile
				return true
			}

			return filter(updateEvent.ObjectNew)
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return filter(deleteEvent.Object)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return filter(genericEvent.Object)
		},
	}
}

func annotationRemoved(e event.UpdateEvent, annotation string) bool {
	_, existsInOld := e.ObjectOld.GetAnnotations()[annotation]
	_, existsInNew := e.ObjectNew.GetAnnotations()[annotation]

	return existsInOld && !existsInNew
}

var reservedNamespaceRegex = regexp.MustCompile(`^(openshift|istio-system)$|^(kube|openshift)-.*$`)

func IsReservedNamespace(namepace string) bool {
	return reservedNamespaceRegex.MatchString(namepace)
}
