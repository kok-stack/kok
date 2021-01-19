package controllers

import (
	"fmt"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	v13 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var installPostModule = ParentModule{
	Name: "install-job",
	Sub:  []Module{installPostJob},
}

var installPostJob = &SubModule{
	getObj: func() Object {
		return &v13.Job{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		out := &v13.Job{}
		out.Name = fmt.Sprintf("%s-install-post", c.Name)
		out.Namespace = c.Namespace
		out.Labels = map[string]string{
			"cluster": c.Name,
			"app":     out.Name,
		}
		out.Spec = v13.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   out.Name,
					Labels: out.Labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  "install-post",
						Image: c.Spec.InitSpec.Image,
						Command: []string{
							"kubectl",
							"--kubeconfig=admin/admin.config",
							"create",
							"clusterrolebinding",
							"cluster-node",
							"--clusterrole=cluster-admin",
							"--user=kubernetes-node",
							"--group=system:node",
						},
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
					}},
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
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		}
		return out
	},
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		job := object.(*v13.Job)
		c.Status.PostInstall.Status = job.Status
		c.Status.PostInstall.Name = job.Name
	},
}
