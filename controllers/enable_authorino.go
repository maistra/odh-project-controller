package controllers

import (
	"context"
	authorino "github.com/kuadrant/authorino/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
	"regexp"
)

type reconcileFunc func(ctx context.Context, ns *v1.Namespace) error

func (r *OpenshiftServiceMeshReconciler) reconcileAuthConfig(ctx context.Context, ns *v1.Namespace) error {
	log := r.Log.WithValues("feature", "authorino", "namespace", ns.Name)

	if ns.Annotations[AnnotationHubURL] == "" {
		log.V(1).Info("Unable to create AuthConfig because of missing annotation", "expected-annotation", AnnotationHubURL)
		return nil
	}

	desiredAuthConfig := r.createAuthConfig(ns)

	foundAuthConfig := &authorino.AuthConfig{}
	justCreated := false
	err := r.Get(ctx, types.NamespacedName{
		Name:      desiredAuthConfig.Name,
		Namespace: ns.Name,
	}, foundAuthConfig)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating Authorino AuthConfig")
			// Create the AuthConfig in the Openshift cluster
			err = r.Create(ctx, desiredAuthConfig)
			if err != nil && !apierrs.IsAlreadyExists(err) {
				log.Error(err, "Unable to create Authorino AuthConfig")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "Unable to fetch the AuthConfig")
			return err
		}
	}

	// Reconcile the Authorino AuthConfig if it has been manually modified
	if !justCreated && !CompareAuthConfigs(*desiredAuthConfig, *foundAuthConfig) {
		log.Info("Reconciling Authorino AuthConfig")
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Get(ctx, types.NamespacedName{
				Name:      desiredAuthConfig.Name,
				Namespace: ns.Name,
			}, foundAuthConfig); err != nil {
				return err
			}
			foundAuthConfig.Spec = desiredAuthConfig.Spec
			foundAuthConfig.ObjectMeta.Labels = desiredAuthConfig.ObjectMeta.Labels
			return r.Update(ctx, foundAuthConfig)
		})
		if err != nil {
			log.Error(err, "Unable to reconcile the Authorino AuthConfig")
			return err
		}
	}

	return nil
}

func (r *OpenshiftServiceMeshReconciler) createAuthConfig(ns *v1.Namespace) *authorino.AuthConfig {
	return &authorino.AuthConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AuthConfig",
			APIVersion: "authorino.kuadrant.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ns.Name + "-protection",
			Namespace: ns.Name,
			Labels: map[string]string{
				"authorino/topic": "odh",
			},
		},
		Spec: authorino.AuthConfigSpec{
			Hosts: []string{
				removeProtocolPrefix(ns.Annotations[AnnotationHubURL]),
			},
			Identity: []*authorino.Identity{
				{
					Name: "authorized-service-accounts",
					KubernetesAuth: &authorino.Identity_KubernetesAuth{
						Audiences: []string{
							"https://kubernetes.default.svc",
						},
					},
				},
			},
			Authorization: []*authorino.Authorization{
				{
					Name: "k8s-rbac",
					KubernetesAuthz: &authorino.Authorization_KubernetesAuthz{
						User: authorino.StaticOrDynamicValue{
							ValueFrom: authorino.ValueFrom{
								AuthJSON: "auth.identity.username",
							},
						},
					},
				},
			},
			Response: []*authorino.Response{
				{
					Name: "x-auth-data",
					JSON: &authorino.Response_DynamicJSON{
						Properties: []authorino.JsonProperty{
							{
								Name: "username",
								ValueFrom: authorino.ValueFrom{
									AuthJSON: "auth.identity.username",
								},
							},
						},
					},
				},
			},
			DenyWith: &authorino.DenyWith{
				Unauthorized: &authorino.DenyWithSpec{
					Message: &authorino.StaticOrDynamicValue{
						Value: "Authorino Denied",
					},
				},
			},
		},
	}
}

func CompareAuthConfigs(a1, a2 authorino.AuthConfig) bool {
	return reflect.DeepEqual(a1.ObjectMeta.Labels, a2.ObjectMeta.Labels) &&
		reflect.DeepEqual(a1.Spec, a2.Spec)
}

func removeProtocolPrefix(s string) string {
	r := regexp.MustCompile(`^(https?://)`)
	return r.ReplaceAllString(s, "")
}
