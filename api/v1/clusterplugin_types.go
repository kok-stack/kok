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

type ClusterPluginPodContainer struct {
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// +optional
	Image string `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	// +optional
	Command []string `json:"command,omitempty" protobuf:"bytes,3,rep,name=command"`
	// +optional
	Args []string `json:"args,omitempty" protobuf:"bytes,4,rep,name=args"`
	// +optional
	WorkingDir string `json:"workingDir,omitempty" protobuf:"bytes,5,opt,name=workingDir"`
	// +optional
	EnvFrom []v1.EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,19,rep,name=envFrom"`
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []v1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,9,rep,name=volumeMounts"`
	// +patchMergeKey=devicePath
	// +patchStrategy=merge
	// +optional
	VolumeDevices []v1.VolumeDevice `json:"volumeDevices,omitempty" patchStrategy:"merge" patchMergeKey:"devicePath" protobuf:"bytes,21,rep,name=volumeDevices"`
	// +optional
	LivenessProbe *v1.Probe `json:"livenessProbe,omitempty" protobuf:"bytes,10,opt,name=livenessProbe"`
	// +optional
	ReadinessProbe *v1.Probe `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`
	// +optional
	StartupProbe *v1.Probe `json:"startupProbe,omitempty" protobuf:"bytes,22,opt,name=startupProbe"`
	// +optional
	Lifecycle *v1.Lifecycle `json:"lifecycle,omitempty" protobuf:"bytes,12,opt,name=lifecycle"`
	// +optional
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`
	// +optional
	SecurityContext *v1.SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,15,opt,name=securityContext"`
}

type ClusterPluginPodSpec struct {
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []v1.Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
	// +patchMergeKey=name
	// +patchStrategy=merge
	InitContainers []ClusterPluginPodContainer `json:"initContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,20,rep,name=initContainers"`
	// +patchMergeKey=name
	// +patchStrategy=merge
	Containers []ClusterPluginPodContainer `json:"containers" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`

	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,8,opt,name=serviceAccountName"`
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,16,opt,name=hostname"`
	// +optional
	RuntimeClassName *string `json:"runtimeClassName,omitempty" protobuf:"bytes,29,opt,name=runtimeClassName"`
}

type CLusterPluginSpecInner struct {
	Install   ClusterPluginPodSpec `json:"install,omitempty"`
	Uninstall ClusterPluginPodSpec `json:"uninstall,omitempty"`
}

// ClusterPluginSpec defines the desired state of ClusterPlugin
type ClusterPluginSpec struct {
	CLusterPluginSpecInner `json:",inline"`
	ClusterName            string `json:"clusterName"`
}

type ClusterPluginPodStatus struct {
	PodName string      `json:"podName,omitempty"`
	Status  v1.PodPhase `json:"status,omitempty"`
}

// ClusterPluginStatus defines the observed state of ClusterPlugin
type ClusterPluginStatus struct {
	InstallStatus   ClusterPluginPodStatus `json:"installStatus,omitempty"`
	UninstallStatus ClusterPluginPodStatus `json:"uninstallStatus,omitempty"`
}

func (in *ClusterPlugin) GetSpec() CLusterPluginSpecInner {
	return in.Spec.CLusterPluginSpecInner
}

func (in *ClusterPlugin) GetStatus() ClusterPluginStatus {
	return in.Status
}

func (in *ClusterPlugin) GetClusterNames() string {
	return in.Spec.ClusterName
}

func (in *ClusterPlugin) UpdateStatus(target ClusterPluginStatus) {
	in.Status = target
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="cluster",type="string",JSONPath=".spec.clusterName",description="cluster_name"
// +kubebuilder:printcolumn:name="install-pod",type="string",JSONPath=".status.installStatus.podName",description="install-pod_name"
// +kubebuilder:printcolumn:name="install-ready",type="string",JSONPath=".status.installStatus.status",description="install-ready"
// +kubebuilder:printcolumn:name="uninstall-pod",type="string",JSONPath=".status.installStatus.podName",description="uninstall-pod_name"
// +kubebuilder:printcolumn:name="uninstall-ready",type="string",JSONPath=".status.uninstallStatus.status",description="uninstall-ready"

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
