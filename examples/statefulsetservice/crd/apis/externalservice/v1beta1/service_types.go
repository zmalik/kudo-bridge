/*

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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExternalServiceSpec defines the desired state of Instance.
type ExternalServiceSpec struct {
	// StatefulSet specifies the reference to Statefulset Object
	StatefulSet corev1.ObjectReference `json:"statefulset,omitempty"`

	Port                  int    `json:"port,omitempty"`
	Count                 int    `json:"count,omitempty"`
	TargetPort            int    `json:"targetPort,omitempty"`
	Type                  string `json:"type,omitempty"`
	ExternalTrafficPolicy string `json:"externalTrafficPolicy,omitempty"`
}

// ExternalServiceStatus defines the observed state of Instance
type ExternalServiceStatus struct {
	Status string `json:"externalServiceStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Instance is the Schema for the instances API.
// +k8s:openapi-gen=true
type ExternalService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExternalServiceSpec   `json:"spec,omitempty"`
	Status ExternalServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ExternalServiceList contains a list of ExternalService.
type ExternalServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalService `json:"items"`
}

func init() {
	SchemeBuilder.Register(AddKnownTypes)
}
