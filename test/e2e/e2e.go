/*
Copyright 2019 The Knative Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacV1beta1 "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing-contrib/test"
	"knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	pkgTest "knative.dev/pkg/test"
	"knative.dev/pkg/test/logging"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"

	// Mysteriously required to support GCP auth (required by k8s libs).
	// Apparently just importing it is enough. @_@ side effects @_@.
	// https://github.com/kubernetes/client-go/issues/242
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	defaultNamespaceName = "e2etest"
	testNamespace        = "e2etest"
	interval             = 1 * time.Second
	timeout              = 1 * time.Minute
)

// Setup creates the client objects needed in the e2e tests.
func Setup(t *testing.T, logger logging.FormatLogger) (*test.Clients, *test.Cleaner) {
	if pkgTest.Flags.Namespace == "" {
		pkgTest.Flags.Namespace = defaultNamespaceName
	}

	clients, err := test.NewClients(
		pkgTest.Flags.Kubeconfig,
		pkgTest.Flags.Cluster,
		pkgTest.Flags.Namespace)
	if err != nil {
		t.Fatalf("Couldn't initialize clients: %v", err)
	}
	cleaner := test.NewCleaner(logger, clients.Dynamic)

	return clients, cleaner
}

// TearDown will delete created names using clients.
func TearDown(clients *test.Clients, cleaner *test.Cleaner, logger logging.FormatLogger) {
	cleaner.Clean(true)
}

// CreateRouteAndConfig will create Route and Config objects using clients.
// The Config object will serve requests to a container started from the image at imagePath.
func CreateRouteAndConfig(clients *test.Clients, logger logging.FormatLogger, cleaner *test.Cleaner, name string, imagePath string) error {
	configurations := clients.Serving.ServingV1beta1().Configurations(pkgTest.Flags.Namespace)
	config, err := configurations.Create(
		test.Configuration(name, pkgTest.Flags.Namespace, imagePath))
	if err != nil {
		return err
	}
	cleaner.Add(servingv1beta1.SchemeGroupVersion.Group, servingv1beta1.SchemeGroupVersion.Version, "configurations", pkgTest.Flags.Namespace, config.ObjectMeta.Name)

	routes := clients.Serving.ServingV1beta1().Routes(pkgTest.Flags.Namespace)
	route, err := routes.Create(
		test.Route(name, pkgTest.Flags.Namespace, name))
	if err != nil {
		return err
	}
	cleaner.Add(servingv1beta1.SchemeGroupVersion.Group, servingv1beta1.SchemeGroupVersion.Version, "routes", pkgTest.Flags.Namespace, route.ObjectMeta.Name)
	return nil
}

// WithRouteReady will create Route and Config objects and wait until they're ready.
func WithRouteReady(clients *test.Clients, logger logging.FormatLogger, cleaner *test.Cleaner, name string, imagePath string) error {
	err := CreateRouteAndConfig(clients, logger, cleaner, name, imagePath)
	if err != nil {
		return err
	}
	routes := clients.Serving.ServingV1beta1().Routes(pkgTest.Flags.Namespace)
	if err := test.WaitForRouteState(routes, name, test.IsRouteReady, "RouteIsReady"); err != nil {
		return err
	}
	return nil
}

// CreateChannel will create a Channel
func CreateChannel(clients *test.Clients, channel *v1alpha1.Channel, logger logging.FormatLogger, cleaner *test.Cleaner) error {
	channels := clients.Eventing.MessagingV1alpha1().Channels(pkgTest.Flags.Namespace)
	res, err := channels.Create(channel)
	if err != nil {
		return err
	}
	cleaner.Add(v1alpha1.SchemeGroupVersion.Group, v1alpha1.SchemeGroupVersion.Version, "channels", pkgTest.Flags.Namespace, res.ObjectMeta.Name)
	return nil
}

// CreateSubscription will create a Subscription
func CreateSubscription(clients *test.Clients, subs *v1alpha1.Subscription, logger logging.FormatLogger, cleaner *test.Cleaner) error {
	subscriptions := clients.Eventing.MessagingV1alpha1().Subscriptions(pkgTest.Flags.Namespace)
	res, err := subscriptions.Create(subs)
	if err != nil {
		return err
	}
	cleaner.Add(v1alpha1.SchemeGroupVersion.Group, v1alpha1.SchemeGroupVersion.Version, "subscriptions", pkgTest.Flags.Namespace, res.ObjectMeta.Name)
	return nil
}

// CreateServiceAccount will create a service account
func CreateServiceAccount(clients *test.Clients, sa *corev1.ServiceAccount, logger logging.FormatLogger, cleaner *test.Cleaner) error {
	sas := clients.Kube.Kube.CoreV1().ServiceAccounts(pkgTest.Flags.Namespace)
	res, err := sas.Create(sa)
	if err != nil {
		return err
	}
	cleaner.Add(corev1.SchemeGroupVersion.Group, corev1.SchemeGroupVersion.Version, "serviceaccounts", pkgTest.Flags.Namespace, res.ObjectMeta.Name)
	return nil
}

// CreateClusterRoleBinding will create a service account binding
func CreateClusterRoleBinding(clients *test.Clients, crb *rbacV1beta1.ClusterRoleBinding, logger logging.FormatLogger, cleaner *test.Cleaner) error {
	clusterRoleBindings := clients.Kube.Kube.RbacV1beta1().ClusterRoleBindings()
	res, err := clusterRoleBindings.Create(crb)
	if err != nil {
		return err
	}
	cleaner.Add(rbacV1beta1.SchemeGroupVersion.Group, rbacV1beta1.SchemeGroupVersion.Version, "clusterrolebindings", "", res.ObjectMeta.Name)
	return nil
}

// CreateServiceAccountAndBinding creates both ServiceAccount and ClusterRoleBinding with default
// cluster-admin role
func CreateServiceAccountAndBinding(clients *test.Clients, name string, logger logging.FormatLogger, cleaner *test.Cleaner) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: pkgTest.Flags.Namespace,
		},
	}
	err := CreateServiceAccount(clients, sa, logger, cleaner)
	if err != nil {
		return err
	}
	crb := &rbacV1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "e2e-tests-admin",
		},
		Subjects: []rbacV1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: pkgTest.Flags.Namespace,
			},
		},
		RoleRef: rbacV1beta1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	err = CreateClusterRoleBinding(clients, crb, logger, cleaner)
	if err != nil {
		return err
	}
	return nil
}

// CreatePod will create a Pod
func CreatePod(clients *test.Clients, pod *corev1.Pod, logger logging.FormatLogger, cleaner *test.Cleaner) error {
	res, err := clients.Kube.CreatePod(pod)
	if err != nil {
		return err
	}

	cleaner.Add(corev1.SchemeGroupVersion.Group, corev1.SchemeGroupVersion.Version, "pods", res.ObjectMeta.Namespace, res.ObjectMeta.Name)
	return nil
}
