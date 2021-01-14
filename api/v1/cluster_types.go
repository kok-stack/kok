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
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ImageBase struct {
	Image string `json:"image"`
}

type ClusterAccessSpec struct {
	Address string `json:"address"`
	Port    string `json:"port"`
}

type ClusterInitSpec struct {
	ImageBase `json:",inline"`
}

type ClusterEtcdSpec struct {
	ImageBase `json:",inline"`
}

type ClusterApiServerSpec struct {
	ImageBase `json:",inline"`
}

type ClusterControllerManagerSpec struct {
	ImageBase `json:",inline"`
}

type ClusterSchedulerSpec struct {
	ImageBase `json:",inline"`
}

type ClusterClientSpec struct {
	ImageBase `json:",inline"`
}

type ClusterKubeletSpec struct {
	PodInfraContainerImage string `json:"podInfraContainerImage"`
}

type ClusterKubeProxySpec struct {
	BindAddress string `json:"bindAddress,omitempty"`
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	//TODO:domain设置
	ClusterVersion        string `json:"clusterVersion"`
	ClusterCIDR           string `json:"clusterCidr"`
	ServiceClusterIpRange string `json:"serviceClusterIpRange"`
	//TODO:image mirror
	AccessSpec ClusterAccessSpec `json:"access"`
	InitSpec   ClusterInitSpec   `json:"init,omitempty"`
	//TODO:count,etcdOperator
	EtcdSpec ClusterEtcdSpec `json:"etcd,omitempty"`
	//TODO:count
	ApiServerSpec ClusterApiServerSpec `json:"apiServer,omitempty"`
	//TODO:count
	ControllerManagerSpec ClusterControllerManagerSpec `json:"controllerManager,omitempty"`
	//TODO:count
	SchedulerSpec ClusterSchedulerSpec `json:"scheduler,omitempty"`
	ClientSpec    ClusterClientSpec    `json:"client,omitempty"`
	KubeletSpec   ClusterKubeletSpec   `json:"kubelet,omitempty"`
	KubeProxySpec ClusterKubeProxySpec `json:"kubeProxy,omitempty"`
}

type ClusterInitStatus struct {
	Name               string            `json:"name,omitempty"`
	CaPkiName          string            `json:"caPkiName,omitempty"`
	EtcdPkiName        string            `json:"etcdPkiName,omitempty"`
	ServerName         string            `json:"serverName,omitempty"`
	ClientName         string            `json:"clientName,omitempty"`
	AdminConfigName    string            `json:"adminConfigName,omitempty"`
	NodeConfigName     string            `json:"nodeConfigName,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	RoleBindingName    string            `json:"roleBindingName,omitempty"`
	Status             batchv1.JobStatus `json:"status,omitempty"`
	DnsAddr            string            `json:"dnsAddr,omitempty"`
}

type ClusterEtcdStatus struct {
	Name    string                  `json:"name,omitempty"`
	SvcName string                  `json:"svcName,omitempty"`
	Status  appsv1.DeploymentStatus `json:"status,omitempty"`
}

type ClusterApiServerStatus struct {
	Name    string                  `json:"name,omitempty"`
	SvcName string                  `json:"svcName,omitempty"`
	Status  appsv1.DeploymentStatus `json:"status,omitempty"`
}

type ClusterControllerManagerStatus struct {
	Name   string                  `json:"name,omitempty"`
	Status appsv1.DeploymentStatus `json:"status,omitempty"`
}

type CLusterSchedulerStatus struct {
	Name   string                  `json:"name,omitempty"`
	Status appsv1.DeploymentStatus `json:"status,omitempty"`
}

type ClusterClientStatus struct {
	Name   string                  `json:"name,omitempty"`
	Status appsv1.DeploymentStatus `json:"status,omitempty"`
}

type ClusterPostInstallStatus struct {
	Name   string            `json:"name,omitempty"`
	Status batchv1.JobStatus `json:"status,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Init              ClusterInitStatus              `json:"init,omitempty"`
	Etcd              ClusterEtcdStatus              `json:"etcd,omitempty"`
	ApiServer         ClusterApiServerStatus         `json:"apiServer,omitempty"`
	ControllerManager ClusterControllerManagerStatus `json:"controllerManager,omitempty"`
	Scheduler         CLusterSchedulerStatus         `json:"scheduler,omitempty"`
	Client            ClusterClientStatus            `json:"client,omitempty"`
	PostInstall       ClusterPostInstallStatus       `json:"postInstall,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="version",type="string",JSONPath=".spec.clusterVersion",description="clusterVersion"
// +kubebuilder:printcolumn:name="cluster-Cidr",type="string",JSONPath=".spec.clusterCidr",description="clusterCidr"
// +kubebuilder:printcolumn:name="cluster-Dns-Addr",type="string",JSONPath=".spec.clusterDnsAddr",description="clusterDnsAddr"
// +kubebuilder:printcolumn:name="service-Cluster-IpRange",type="string",JSONPath=".spec.serviceClusterIpRange",description="serviceClusterIpRange"
// +kubebuilder:printcolumn:name="access-address",type="string",JSONPath=".spec.access.address",description="access-address"
// +kubebuilder:printcolumn:name="access-port",type="string",JSONPath=".spec.access.port",description="access-port"

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
