/*
Copyright 2022.

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

// WebserverSpec defines the desired state of Webserver
type WebserverSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The number of replicas that the webserver should have
	Replicas int `json:"replicas,omitempty"`

	// The container image of the webserver
	Image string `json:"image,omitempty"`
	// Foo is an example field of Webserver. Edit webserver_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// WebserverStatus defines the observed state of Webserver
type WebserverStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Webserver is the Schema for the webservers API
type Webserver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebserverSpec   `json:"spec,omitempty"`
	Status WebserverStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WebserverList contains a list of Webserver
type WebserverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webserver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Webserver{}, &WebserverList{})
}
