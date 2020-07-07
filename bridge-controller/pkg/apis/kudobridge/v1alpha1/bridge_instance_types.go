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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// BridgeInstanceSpec defines the desired state of Instance.
type BridgeInstanceSpec struct {
	// KUDOOperator specifies the KUDO Operator
	KUDOOperator KUDOOperator `json:"kudoOperator,omitempty"`

	//CRDSpec specifies the CRD to watch
	CRDSpec unstructured.Unstructured `json:"crdSpec,omitempty"`
}

// KUDOOperator defines the KUDO Operator reference definition
type KUDOOperator struct {
	//Package specifies the KUDO package name
	Package string `json:"package,omitempty"`
	//KUDORepository specifies the KUDO Repository URL
	KUDORepository string `json:"repository,omitempty"`
	//InClusterOperator is used to resolve incluster operator
	InClusterOperator bool `json:"inClusterOperator,omitempty"`
	//Version specifies the KUDO Operator Version
	Version string `json:"version,omitempty"`
	//AppVersion specifies the KUDO Operator Application Version
	AppVersion string `json:"appVersion,omitempty"`
}

// BridgeInstanceStatus defines the observed state of Instance
type BridgeInstanceStatus struct {
	Status string `json:"bridgeInstanceStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Instance is the Schema for the instances API.
// +k8s:openapi-gen=true
type BridgeInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BridgeInstanceSpec   `json:"spec,omitempty"`
	Status BridgeInstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// BridgeInstanceList contains a list of BridgeInstance.
type BridgeInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BridgeInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(AddKnownTypes)
}
