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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterPluginSpec defines the desired state of ClusterPlugin
type ClusterPluginSpec struct {
	ClusterName string     `json:"cluster_name"`
	Install     v1.PodSpec `json:"install"`
	Uninstall   v1.PodSpec `json:"uninstall"`
}

type ClusterPluginPodStatus struct {
	Ready  bool         `json:"ready"`
	Status v1.PodStatus `json:"status"`
}

// ClusterPluginStatus defines the observed state of ClusterPlugin
type ClusterPluginStatus struct {
	InstallStatus   ClusterPluginPodStatus `json:"install_status"`
	UninstallStatus ClusterPluginPodStatus `json:"uninstall_status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="cluster",type="string",JSONPath=".spec.cluster_name",description="cluster_name"
// +kubebuilder:printcolumn:name="install-image",type="string",JSONPath=".spec.install.containers.image",description="install-image"
// +kubebuilder:printcolumn:name="install-ready",type="string",JSONPath=".status.install_status.ready",description="install-ready"
// +kubebuilder:printcolumn:name="uninstall-image",type="string",JSONPath=".spec.uninstall.containers.image",description="uninstall-image"
// +kubebuilder:printcolumn:name="uninstall-ready",type="string",JSONPath=".status.uninstall_status.ready",description="uninstall-ready"

// ClusterPlugin is the Schema for the clusterplugins API
type ClusterPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterPluginSpec   `json:"spec,omitempty"`
	Status ClusterPluginStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterPluginList contains a list of ClusterPlugin
type ClusterPluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPlugin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterPlugin{}, &ClusterPluginList{})
}
