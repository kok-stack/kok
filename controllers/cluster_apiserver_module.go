package controllers

import (
	"fmt"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var apiServerModule = ParentModule{
	Name: "apiserver-dept",
	Sub:  []Module{apiServerDept, apiServerSvc},
}

var apiServerDept = &SubModule{
	getObj: func() Object {
		return &v12.Deployment{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		//TODO:定制Replicas数量
		var rep int32 = 1
		name := fmt.Sprintf("%s-apiserver", c.Name)
		var out = &v12.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: c.Namespace,
			},
			Spec: v12.DeploymentSpec{
				Replicas: &rep,
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: c.Namespace,
						Labels: map[string]string{
							"cluster": c.Name,
							"app":     name,
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "apiserver",
								Image: c.Spec.ApiServerSpec.Image,
								Command: []string{
									"kube-apiserver",
									"--allow-privileged=true",
									"--authorization-mode=Node,RBAC",
									"--client-ca-file=/pki/ca/ca.pem",
									"--enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,TaintNodesByCondition,Priority,DefaultTolerationSeconds,DefaultStorageClass,StorageObjectInUseProtection,PersistentVolumeClaimResize,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,RuntimeClass,ResourceQuota",
									"--etcd-cafile=/pki/ca/ca.pem",
									"--etcd-certfile=/pki/etcd/etcd-client.crt",
									"--etcd-keyfile=/pki/etcd/etcd-client.key",
									fmt.Sprintf("--etcd-servers=https://%s:%v", c.Status.Etcd.SvcName, c.Status.Etcd.Status.ClientPort),
									"--insecure-port=0",
									"--kubelet-client-certificate=/pki/client/kubernetes-node.pem",
									"--kubelet-client-key=/pki/client/kubernetes-node-key.pem",
									"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
									"--secure-port=6443",
									fmt.Sprintf("--service-cluster-ip-range=%s", c.Spec.ServiceClusterIpRange),
									"--tls-cert-file=/pki/server/kubernetes-server.pem",
									"--tls-private-key-file=/pki/server/kubernetes-server-key.pem",
								},
								Ports: []v1.ContainerPort{{
									Name:          "https-6443",
									ContainerPort: 6443,
								}},
								VolumeMounts: []v1.VolumeMount{
									{
										Name:      "ca-pki",
										ReadOnly:  true,
										MountPath: "/pki/ca",
									},
									{
										Name:      "etcd-pki",
										ReadOnly:  true,
										MountPath: "/pki/etcd",
									},
									{
										Name:      "k8s-server",
										ReadOnly:  true,
										MountPath: "/pki/server",
									},
									{
										Name:      "k8s-client",
										ReadOnly:  true,
										MountPath: "/pki/client",
									},
								},
							},
						},
						Volumes: []v1.Volume{{
							Name: "ca-pki",
							VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
								SecretName: c.Status.Init.CaPkiName,
							}},
						}, {
							Name: "etcd-pki",
							VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
								SecretName: c.Status.Init.EtcdPkiClientName,
							}},
						}, {
							Name: "k8s-server",
							VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
								SecretName: c.Status.Init.ServerName,
							}},
						}, {
							Name: "k8s-client",
							VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
								SecretName: c.Status.Init.ClientName,
							}},
						},
						},
					},
				},
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster": c.Name,
						"app":     name,
					},
				},
			},
		}
		return out
	},
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		dept := object.(*v12.Deployment)
		c.Status.ApiServer.Status = dept.Status
		c.Status.ApiServer.Name = dept.Name
	},
}

var apiServerSvc = &SubModule{
	getObj: func() Object {
		return &v1.Service{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		name := fmt.Sprintf("%s-apiserver", c.Name)
		out := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: c.Namespace,
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{
					"cluster": c.Name,
					"app":     name,
				},
				Type: v1.ServiceTypeNodePort,
				Ports: []v1.ServicePort{
					{
						Name: "https-6443",
						Port: 6443,
					},
				},
			},
		}
		return out
	},
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		svc := object.(*v1.Service)
		c.Status.ApiServer.SvcName = svc.Name
	},
	ready: func(c *tanxv1.Cluster) bool {
		for _, condition := range c.Status.ApiServer.Status.Conditions {
			if v12.DeploymentAvailable == condition.Type {
				return true
			}
		}
		return false
	},
}
