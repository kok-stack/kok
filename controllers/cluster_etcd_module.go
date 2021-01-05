package controllers

import (
	"fmt"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var etcdModule = ParentModule{
	Sub: []Module{etcdDept, etcdSvc},
}

var etcdDept = &SubModule{
	getObj: func() Object {
		return &v12.Deployment{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		var rep int32 = 1
		name := fmt.Sprintf("%s-etcd", c.Name)
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
								Name:  "etcd",
								Image: c.Spec.EtcdSpec.Image,
								Command: []string{
									"etcd",
									"--name",
									"etcd1",
									"--cert-file",
									"/pki/etcd/etcd.pem",
									"--key-file",
									"/pki/etcd/etcd-key.pem",
									"--peer-cert-file",
									"/pki/etcd/etcd.pem",
									"--peer-key-file",
									"/pki/etcd/etcd-key.pem",
									"--trusted-ca-file",
									"/pki/ca/ca.pem",
									"--peer-trusted-ca-file",
									"/pki/ca/ca.pem",
									"--listen-client-urls",
									"http://0.0.0.0:2379",
									"--advertise-client-urls",
									"http://0.0.0.0:2379",
									"--initial-cluster-token",
									"etcd-cluster-token",
									"--initial-cluster-state",
									"new",
									"--data-dir",
									"./etcd-dat",
								},
								Ports: []v1.ContainerPort{{
									Name:          "grpc-2379",
									ContainerPort: 2379,
								}, {
									Name:          "grpc-2380",
									ContainerPort: 2380,
								}},
								Resources: v1.ResourceRequirements{},
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
								SecretName: c.Status.Init.EtcdPkiName,
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
		c.Status.Etcd.Status = dept.Status
		c.Status.Etcd.Name = dept.Name
	},
}

var etcdSvc = &SubModule{
	getObj: func() Object {
		return &v1.Service{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		name := fmt.Sprintf("%s-etcd", c.Name)
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
				Ports: []v1.ServicePort{
					{
						Name: "grpc-2379",
						Port: 2379,
					}, {
						Name: "grpc-2380",
						Port: 2380,
					},
				},
			},
		}
		return out
	},
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		svc := object.(*v1.Service)
		c.Status.Etcd.SvcName = svc.Name
	},
}
