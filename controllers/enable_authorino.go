package controllers

import (
	"context"
	_ "embed" // needed for go:embed directive
	"reflect"
	"regexp"
	"strings"

	authorino "github.com/kuadrant/authorino/api/v1beta1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

type reconcileFunc func(ctx context.Context, namespace *v1.Namespace) error

func (r *OpenshiftServiceMeshReconciler) reconcileAuthConfig(ctx context.Context, namespace *v1.Namespace) error {
	log := r.Log.WithValues("feature", "authorino", "namespace", namespace.Name)

	desiredAuthConfig, err := r.createAuthConfig(namespace,
		namespace.ObjectMeta.Annotations[AnnotationProjectModelGatewayHostPatternExternal],
		namespace.ObjectMeta.Annotations[AnnotationProjectModelGatewayHostPatternInternal])
	if err != nil {
		log.Error(err, "Failed creating AuthConfig object")

		return err
	}

	justCreated := false
	foundAuthConfig := &authorino.AuthConfig{}

	err = r.Get(ctx, types.NamespacedName{
		Name:      desiredAuthConfig.Name,
		Namespace: namespace.Name,
	}, foundAuthConfig)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating Authorino AuthConfig")

			err = r.Create(ctx, desiredAuthConfig)
			if err != nil && !apierrs.IsAlreadyExists(err) {
				log.Error(err, "Unable to create Authorino AuthConfig")

				return errors.Wrap(err, "unable to create AuthConfig")
			}

			justCreated = true
		} else {
			log.Error(err, "Unable to fetch the AuthConfig")

			return errors.Wrap(err, "unable to fetch AuthConfig")
		}
	}

	// Reconcile the Authorino AuthConfig if it has been manually modified
	if !justCreated && !CompareAuthConfigs(*desiredAuthConfig, *foundAuthConfig) {
		log.Info("Reconciling Authorino AuthConfig")

		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Get(ctx, types.NamespacedName{
				Name:      desiredAuthConfig.Name,
				Namespace: namespace.Name,
			}, foundAuthConfig); err != nil {
				return errors.Wrapf(err, "failed getting AuthConfig %s in namespace %s", desiredAuthConfig.Name, namespace.Name)
			}

			foundAuthConfig.Spec = desiredAuthConfig.Spec
			foundAuthConfig.ObjectMeta.Labels = desiredAuthConfig.ObjectMeta.Labels

			return errors.Wrap(r.Update(ctx, foundAuthConfig), "failed updating AuthConfig")
		}); err != nil {
			log.Error(err, "Unable to reconcile the Authorino AuthConfig")

			return errors.Wrap(err, "unable to reconcile the Authorino AuthConfig")
		}
	}

	return nil
}

//go:embed template/authconfig.yaml
var authConfigTemplate []byte

func (r *OpenshiftServiceMeshReconciler) createAuthConfig(namespace *v1.Namespace, hosts ...string) (*authorino.AuthConfig, error) {
	authHosts := make([]string, len(hosts))
	for i := range hosts {
		authHosts[i] = ExtractHostName(hosts[i])
	}

	authConfig := &authorino.AuthConfig{}
	if err := ConvertToStructuredResource(authConfigTemplate, authConfig); err != nil {
		return authConfig, err
	}

	authConfig.SetName(namespace.Name + "-protection")
	authConfig.SetNamespace(namespace.Name)

	keyValue, err := getAuthorinoLabel()
	if err != nil {
		return nil, err
	}

	authConfig.Labels[keyValue[0]] = keyValue[1]
	authConfig.Spec.Hosts = authHosts
	authConfig.Spec.Identity[0].KubernetesAuth.Audiences = []string{namespace.Name + "-api"}

	return authConfig, nil
}

func CompareAuthConfigs(a1, a2 authorino.AuthConfig) bool {
	return reflect.DeepEqual(a1.ObjectMeta.Labels, a2.ObjectMeta.Labels) &&
		reflect.DeepEqual(a1.Spec, a2.Spec)
}

// ExtractHostName strips given URL in string from http(s):// prefix and subsequent path.
// This is useful when getting value from http headers (such as origin), as Authorino needs host only.
// If given string does not start with http(s) prefix it will be returned as is.
func ExtractHostName(s string) string {
	r := regexp.MustCompile(`^(https?://)`)

	withoutProtocol := r.ReplaceAllString(s, "")
	if s == withoutProtocol {
		return s
	}

	index := strings.Index(withoutProtocol, "/")
	if index == -1 {
		return withoutProtocol
	}

	return withoutProtocol[:index]
}
