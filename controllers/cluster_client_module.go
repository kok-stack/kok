package controllers

import (
	"fmt"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

var clientModule = &Module{
	Name: "client-dept",
	Sub:  []*Module{clientDept},
}

var clientDept = &Module{
	getObj: func() Object {
		return &v12.Deployment{}
	},
	render: func(c *tanxv1.Cluster) Object {
		var rep int32 = 1
		var termination int64 = 1
		name := fmt.Sprintf("%s-client", c.Name)
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
						TerminationGracePeriodSeconds: &termination,
						Containers: []v1.Container{
							{
								Name:  "scheduler",
								Image: c.Spec.ClientSpec.Image,
								Command: []string{
									"cat",
								},
								Stdin: true,
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
									{
										Name:      "nodeconfig",
										ReadOnly:  true,
										MountPath: "/home/node",
									},
									{
										Name:      "adminconfig",
										ReadOnly:  true,
										MountPath: "/home/admin",
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
						}, {
							Name: "nodeconfig",
							VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
								SecretName: c.Status.Init.NodeConfigName,
							}},
						}, {
							Name: "adminconfig",
							VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
								SecretName: c.Status.Init.AdminConfigName,
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
	setStatus: func(c *tanxv1.Cluster, target, now Object) (bool, Object) {
		dept := now.(*v12.Deployment)
		c.Status.Client.Status = dept.Status
		c.Status.Client.Name = dept.Name

		t := target.(*v12.Deployment)
		n := now.(*v12.Deployment)
		if !reflect.DeepEqual(t.Spec, n.Spec) {
			n.Spec = t.Spec
			return true, n
		}
		return false, n
	},
}
