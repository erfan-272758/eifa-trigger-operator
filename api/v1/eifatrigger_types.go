/*
Copyright 2025 Erfan.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	FAILED  = "Failed"
	SUCCESS = "Success"
)

// EifaTriggerSpec defines the desired state of EifaTrigger
type WatchSelector struct {
	// +kubebuilder:validation:Enum={"ConfigMap","Secret"}
	Kind          string            `json:"kind"`
	LabelSelector map[string]string `json:"labelSelector"`
}
type UpdateSelector struct {
	// +kubebuilder:validation:Enum={"Deployment","DaemonSet"}
	Kind          string            `json:"kind"`
	LabelSelector map[string]string `json:"labelSelector"`
}
type EifaTriggerSpec struct {
	Watch      WatchSelector    `json:"watch,omitempty"`
	Update     UpdateSelector   `json:"update,omitempty"`
	WatchList  []WatchSelector  `json:"watchList,omitempty"`
	UpdateList []UpdateSelector `json:"updateList,omitempty"`
}

// EifaTriggerStatus defines the observed state of EifaTrigger
type EifaTriggerStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=et

// EifaTrigger is the Schema for the eifatriggers API
type EifaTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:XValidation:rule="has(self.update.kind) || size(self.updateList) > 0",message="Either 'update' or 'updateList' must be specified"
	// +kubebuilder:validation:XValidation:rule="has(self.watch.kind) || size(self.watchList) > 0",message="Either 'watch' or 'watchList' must be specified"
	Spec   EifaTriggerSpec   `json:"spec,omitempty"`
	Status EifaTriggerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// EifaTriggerList contains a list of EifaTrigger
type EifaTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EifaTrigger `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EifaTrigger{}, &EifaTriggerList{})
}
