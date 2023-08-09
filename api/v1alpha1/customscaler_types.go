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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CustomScalerSpec struct {
	// DeploymentName is the name of the deployment to be scaled.
	DeploymentName string `json:"deploymentName"`

	// MetricSource denotes the source of the custom metric to monitor.
	MetricSource string `json:"metricSource"`

	// ScaleUpThreshold is the value above which the deployment should scale up. e.g 80
	ScaleUpThreshold int32 `json:"scaleUpThreshold"`

	// ScaleDownThreshold is the value below which the deployment should scale down. e.g 20
	ScaleDownThreshold int32 `json:"scaleDownThreshold"`

	// CooldownPeriod denotes the amount of time (in seconds) to wait between scaling actions.
	CooldownPeriod int32 `json:"cooldownPeriod"`

	// MaxReplicas is the maximum number of replicas the deployment can have.
	MaxReplicas int32 `json:"maxReplicas"`

	// MinReplicas is the minimum number of replicas the deployment should maintain.
	MinReplicas int32 `json:"minReplicas"`
}

// CustomScalerStatus defines the observed state of CustomScaler
type CustomScalerStatus struct {
	// CurrentReplicas captures the current number of replicas of the deployment.
	CurrentReplicas int32 `json:"currentReplicas"`

	// LastScaledTimestamp captures the last time the deployment was scaled.
	LastScaleTime metav1.Time `json:"lastScaledTime"`

	// CurrentMetricValue captures the current value of the custom metric.
	CurrentMetricValue int32 `json:"currentMetricValue"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CustomScaler is the Schema for the customscalers API
type CustomScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomScalerSpec   `json:"spec,omitempty"`
	Status CustomScalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CustomScalerList contains a list of CustomScaler
type CustomScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomScaler{}, &CustomScalerList{})
}
