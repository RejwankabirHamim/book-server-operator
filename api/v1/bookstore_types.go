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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Bookstore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   BookstoreSpec   `json:"spec"`
	Status BookstoreStatus `json:"status,omitempty"`
}

type BookstoreStatus struct {
	State             string `json:"state,omitempty"`
	DeploymenCreated  bool   `json:"deploymenCreated,omitempty"`
	DeploymentMessage string `json:"deploymentMessage,omitempty"`
	ServiceCreated    bool   `json:"serviceCreated,omitempty"`
	ServiceMessage    string `json:"serviceMessage,omitempty"`
}

// +kubebuilder:object:root=true

type BookstoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Bookstore `json:"items"`
}

type BookstoreSpec struct {
	Replicas  *int32        `json:"replicas"`
	Container ContainerSpec `json:"container,container"`
}

// ContainerSpec contains specs of container
type ContainerSpec struct {
	Image string `json:"image,omitempty"`
	Port  int32  `json:"port,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Bookstore{}, &BookstoreList{})
}
