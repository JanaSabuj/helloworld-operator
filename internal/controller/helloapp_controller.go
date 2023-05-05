/*
Copyright 2023.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	customv1 "github.com/JanaSabuj/helloworld-operator/api/v1"
)

// HelloAppReconciler reconciles a HelloApp object
type HelloAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=custom.janasabuj.github.io,resources=helloapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=custom.janasabuj.github.io,resources=helloapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=custom.janasabuj.github.io,resources=helloapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HelloApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *HelloAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lg := log.FromContext(ctx)
	lg.WithValues("HelloApp created: ", req.NamespacedName)

	// Get the CRD
	var helloApp customv1.HelloApp
	if err := r.Get(ctx, req.NamespacedName, &helloApp); err != nil {
		lg.Error(err, "Unable to fetch HelloApp CRD")
		return ctrl.Result{}, nil
	}

	// Check if existing deployment already there
	currentDeployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, currentDeployment); err != nil {
		// No existing deployment
		lg.Info("Deployment not found for CRD!", "err msg", err.Error())
		expectedDeployment := deriveExpectedDeploymentFromCRD(&helloApp, nil)

		lg.Info("Creating a new Deployment", "Namespace: ", expectedDeployment.Namespace, "Name: ", expectedDeployment.Name)
		if err := r.Create(ctx, expectedDeployment); err != nil {
			lg.Error(err, "Failed to create Deployment", "Namespace: ", expectedDeployment.Namespace, "Name: ", expectedDeployment.Name)
			return ctrl.Result{}, err
		}
	} else {
		// Update the existing deployment iff it needs updating
		// if helloApp.Spec.Size == *currentDeployment.Spec.Replicas {
		// 	return ctrl.Result{}, nil
		// }

		updatedDeployment := deriveExpectedDeploymentFromCRD(&helloApp, currentDeployment)
		lg.Info("Updating a new Deployment", "Namespace: ", updatedDeployment.Namespace, "Name: ", updatedDeployment.Name)
		if err := r.Update(ctx, updatedDeployment); err != nil {
			lg.Error(err, "Failed to update Deployment", "Namespace: ", updatedDeployment.Namespace, "Name: ", updatedDeployment.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&customv1.HelloApp{}).
		Complete(r)
}

func deriveExpectedDeploymentFromCRD(helloApp *customv1.HelloApp, currentDeployment *appsv1.Deployment) *appsv1.Deployment {
	newDeployment := &appsv1.Deployment{}
	// if currentDeployment != nil {
	// 	// update and return Dep
	// 	currentDeployment.DeepCopyInto(newDeployment)
	// 	*newDeployment.Spec.Replicas = helloApp.Spec.Size
	// 	return newDeployment
	// }

	// Image
	const PodImage string = "luksa/kubia"

	// Create a Deployment with the same name and replica size as the HelloApp CRD
	newDeployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helloApp.Name,
			Namespace: helloApp.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &helloApp.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": helloApp.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": helloApp.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  helloApp.Name,
							Image: PodImage,
							Ports: []corev1.ContainerPort{
								{
									Name:          "tcp",
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	return newDeployment
}
