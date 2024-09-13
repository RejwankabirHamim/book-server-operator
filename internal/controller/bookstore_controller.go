/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bookscomv1 "books.com/bookstore/api/v1"
)

// BookstoreReconciler reconciles a Bookstore object
type BookstoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=books.com.books.com,resources=bookstores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=books.com.books.com,resources=bookstores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=books.com.books.com,resources=bookstores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bookstore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile

func isDeploymentChanged(bookstore *bookscomv1.Bookstore, deployment *appsv1.Deployment) bool {
	if *bookstore.Spec.Replicas != *deployment.Spec.Replicas {
		*deployment.Spec.Replicas = *bookstore.Spec.Replicas
		return true
	}
	if bookstore.Spec.Container.Image != deployment.Spec.Template.Spec.Containers[0].Image {
		deployment.Spec.Template.Spec.Containers[0].Image = bookstore.Spec.Container.Image
		return true
	}
	return false
}

func isServiceChanged(bookstore *bookscomv1.Bookstore, service *corev1.Service) bool {
	if intstr.FromInt32(bookstore.Spec.Container.Port) != service.Spec.Ports[0].TargetPort {
		service.Spec.Ports[0].TargetPort = intstr.FromInt32(bookstore.Spec.Container.Port)
		return true
	}
	return false
}

func (r *BookstoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fmt.Println("<<<<< In Reconcile Function >>>> ")
	fmt.Println("Name: ", req.NamespacedName.Name, "\nNamespace: ", req.NamespacedName.Namespace)

	// slow down the execution to see the changes clearly
	time.Sleep(time.Second * 5)
	// TODO(user): your logic here

	var bs bookscomv1.Bookstore

	if err := r.Get(ctx, req.NamespacedName, &bs); err != nil {
		fmt.Println("Unable to get Bookstore resource", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	fmt.Println("Checking Deployment...")

	var deployment appsv1.Deployment

	deploymentNameSpace := types.NamespacedName{
		Namespace: req.NamespacedName.Namespace,
		Name:      req.NamespacedName.Name,
	}

	fmt.Println("dep: ", deploymentNameSpace)
	if err := r.Get(ctx, deploymentNameSpace, &deployment); err != nil {
		if !errors.IsNotFound(err) {
			fmt.Println("Deployment is not deleted, some others error")
			return ctrl.Result{}, err
		}
		deployment = *newDeployment(&bs)
		if err := r.Create(ctx, &deployment); err != nil {
			fmt.Println("Deployment creation error", err)
			return ctrl.Result{}, err
		}
		fmt.Println("Deployment created")
	}

	if isDeploymentChanged(&bs, &deployment) {
		fmt.Println("Deployment changed")
		if err := r.Update(ctx, &deployment); err != nil {
			fmt.Println("Deployment update error", err)
			return ctrl.Result{}, err
		}
		fmt.Println("Deployment updated")
	}

	fmt.Println("Checking Service...")
	var svc corev1.Service
	svcNameSpace := types.NamespacedName{
		Namespace: req.NamespacedName.Namespace,
		Name:      req.NamespacedName.Name + "-service",
	}

	fmt.Println("svc: ", svcNameSpace)
	if err := r.Get(ctx, svcNameSpace, &svc); err != nil {
		if !errors.IsNotFound(err) {
			fmt.Println("Service is not deleted, some others error")
			return ctrl.Result{}, err
		}
		svc = *newService(&bs)
		if err := r.Create(ctx, &svc); err != nil {
			fmt.Println("Service creation error", err)
			return ctrl.Result{}, err
		}
		fmt.Println("Service created")
	}

	if isServiceChanged(&bs, &svc) {
		fmt.Println("Service changed")
		if err := r.Update(ctx, &svc); err != nil {
			fmt.Println("Service update error", err)
			return ctrl.Result{}, err
		}
		fmt.Println("Service updated")
	}

	return ctrl.Result{}, nil
}

func newDeployment(bookStore *bookscomv1.Bookstore) *appsv1.Deployment {
	fmt.Println("Inside newDeployment +++++++++++")

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bookStore.ObjectMeta.Name,
			Namespace: bookStore.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(bookStore, bookscomv1.GroupVersion.WithKind("Bookstore")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: bookStore.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "bookserver",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "bookserver",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "bookapi",
							Image: bookStore.Spec.Container.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: bookStore.Spec.Container.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}

func newService(bookStore *bookscomv1.Bookstore) *corev1.Service {
	fmt.Println("---+++--- ", bookStore.Spec.Container.Port)
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bookStore.ObjectMeta.Name + "-service",
			Namespace: bookStore.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(bookStore, bookscomv1.GroupVersion.WithKind("Bookstore")),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": "bookserver",
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       4000,
					TargetPort: intstr.FromInt32(bookStore.Spec.Container.Port),
				},
			},
		},
	}
}

var (
	OwnerKey = ".metadata.controller"
	apiGVStr = bookscomv1.GroupVersion.String()
	ourKind  = "Bookstore"
)

// SetupWithManager sets up the controller with the Manager.
func (r *BookstoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	fmt.Println("...... In Manager ......")

	// for deployment
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, OwnerKey, func(object client.Object) []string {
		// Get the Deployment object
		deployment := object.(*appsv1.Deployment)
		// Extract the owner
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		// Make sure it is a Custom Resource
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		// ...and if so, return it.
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, OwnerKey, func(object client.Object) []string {
		service := object.(*corev1.Service)
		owner := metav1.GetControllerOf(service)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&bookscomv1.Bookstore{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
