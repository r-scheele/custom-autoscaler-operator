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
	"fmt"
	"log"
	"time"

	autoscalev1 "github.com/r-scheele/custom-autoscaler-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// CustomScalerReconciler reconciles a CustomScaler object
type CustomScalerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=autoscale.example.com,resources=customscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscale.example.com,resources=customscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=autoscale.example.com,resources=customscalers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *CustomScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the CustomScaler instance
	customScaler := &autoscalev1.CustomScaler{}
	err := r.Get(ctx, req.NamespacedName, customScaler)
	if err != nil {
		if errors.IsNotFound(err) {
			// If the CustomScaler is not found, return and don't requeue
			return ctrl.Result{}, nil
		}
		// For other errors, requeue with some backoff
		return ctrl.Result{}, err
	}

	// Before scaling up or down:
	lastScaleTime := customScaler.Status.LastScaleTime
	if time.Since(lastScaleTime.Time) < time.Duration(customScaler.Spec.CooldownPeriod)*time.Second {
		return ctrl.Result{RequeueAfter: time.Duration(customScaler.Spec.CooldownPeriod) * time.Second}, nil
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create clientset: %v", err)
	}

	// Fetch the current value of the custom metric
	currentMetricValue, err := fetchMetricValue(customScaler.Spec.MetricSource)
	if err != nil {
		// If there was an error fetching the metric, requeue with some backoff
		return ctrl.Result{}, err
	}

	log.Println("Current metric value: ", currentMetricValue)
	log.Println("Scale up threshold: ", customScaler.Spec.ScaleUpThreshold)
	log.Println("Scale down threshold: ", customScaler.Spec.ScaleDownThreshold)

	// Decide whether to scale up, scale down, or do nothing
	if currentMetricValue > customScaler.Spec.ScaleUpThreshold {
		scaleUp(clientset, customScaler)
	} else if currentMetricValue < customScaler.Spec.ScaleDownThreshold {
		scaleDown(clientset, customScaler)
	}

	now := metav1.Now()
	customScaler.Status.LastScaleTime = now
	// Use the Status().Update() method to update the status subresource
	if err := r.Status().Update(ctx, customScaler); err != nil {
		return ctrl.Result{}, err
	}

	// If everything went fine, don't requeue
	return ctrl.Result{}, nil
}

func getReplicaCount(clientset *kubernetes.Clientset, namespace string, deploymentName string) (int32, error) {

	// Fetch the Deployment object by name
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	// Return the number of replicas. If the replica count is nil, return 0
	if deployment.Spec.Replicas != nil {
		return *deployment.Spec.Replicas, nil
	}
	return 0, nil
}

func scaleUp(clientset *kubernetes.Clientset, customScaler *autoscalev1.CustomScaler) {

	currentReplicas, err := getReplicaCount(clientset, customScaler.Namespace, customScaler.Spec.DeploymentName)
	if err != nil {
		log.Printf("Failed to get replica count: %v", err)
		return
	}

	desiredReplicas := currentReplicas + 1

	// Ensure we don't scale beyond the max allowed replicas
	if desiredReplicas > customScaler.Spec.MaxReplicas {
		desiredReplicas = customScaler.Spec.MaxReplicas
	}

	if desiredReplicas > currentReplicas {
		updateDeploymentReplicas(clientset, customScaler.Namespace, customScaler.Spec.DeploymentName, desiredReplicas)
	}
}

func scaleDown(clientset *kubernetes.Clientset, customScaler *autoscalev1.CustomScaler) {

	currentReplicas, err := getReplicaCount(clientset, customScaler.Namespace, customScaler.Spec.DeploymentName)
	if err != nil {
		log.Printf("Failed to get replica count: %v", err)
		return
	}
	desiredReplicas := currentReplicas - 1

	// Ensure we don't scale below the min allowed replicas
	if desiredReplicas < customScaler.Spec.MinReplicas {
		desiredReplicas = customScaler.Spec.MinReplicas
	}

	if desiredReplicas < currentReplicas {
		updateDeploymentReplicas(clientset, customScaler.Namespace, customScaler.Spec.DeploymentName, desiredReplicas)
	}
}

func updateDeploymentReplicas(clientset *kubernetes.Clientset, customScalerNamespace string, deploymentName string, desiredReplicas int32) error {

	// Get the current deployment
	deploymentsClient := clientset.AppsV1().Deployments(customScalerNamespace)
	deployment, err := deploymentsClient.Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return fmt.Errorf("deployment not found: %v", err)
	} else if err != nil {
		return fmt.Errorf("failed to get deployment: %v", err)
	}

	// Update replica count and patch the deployment
	deployment.Spec.Replicas = &desiredReplicas
	_, err = deploymentsClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment replicas: %v", err)
	}

	return nil
}

func fetchMetricValue(metricSource string) (int32, error) {
	// TODO: Logic to fetch the metric value from the metric source (e.g., Prometheus)

	// Check if metric source is reachable or valid
	if !isMetricSourceValid(metricSource) {
		return 0, errors.NewBadRequest("invalid metric source")
	}
	// Placeholder logic; replace this with actual code to fetch the metric
	return 30, nil
}

func isMetricSourceValid(metricSource string) bool {
	// TODO: Logic to validate the metric source, e.g., making a test connection or validation

	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalev1.CustomScaler{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
