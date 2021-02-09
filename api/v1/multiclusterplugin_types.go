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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MultiClusterPluginSpec defines the desired state of MultiClusterPlugin
type MultiClusterPluginSpec struct {
	Clusters               []string `json:"clusters"`
	CLusterPluginSpecInner `json:",inline"`
}

// +kubebuilder:object:generate=false

type ClusterPluginObj interface {
	metav1.Object
	runtime.Object

	GetSpec() CLusterPluginSpecInner
	GetStatus() ClusterPluginStatus
	GetClusterNames() string
	UpdateStatus(ClusterPluginStatus)
}

// +kubebuilder:object:root=true

// MultiClusterPlugin is the Schema for the multiclusterplugins API
type MultiClusterPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultiClusterPluginSpec `json:"spec,omitempty"`
	Status ClusterPluginStatus    `json:"status,omitempty"`
}

func (in *MultiClusterPlugin) GetSpec() CLusterPluginSpecInner {
	return in.Spec.CLusterPluginSpecInner
}

func (in *MultiClusterPlugin) GetStatus() ClusterPluginStatus {
	return in.Status
}

func (in *MultiClusterPlugin) GetClusterNames() string {
	return strings.Join(in.Spec.Clusters, ",")
}

func (in *MultiClusterPlugin) UpdateStatus(target ClusterPluginStatus) {
	in.Status = target
}

// +kubebuilder:object:root=true

// MultiClusterPluginList contains a list of MultiClusterPlugin
type MultiClusterPluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultiClusterPlugin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MultiClusterPlugin{}, &MultiClusterPluginList{})
}
