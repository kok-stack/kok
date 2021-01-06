package controllers

import (
	"context"
	"fmt"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	v1 "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
	v13 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var initModule = ParentModule{
	Sub: []Module{InitServiceAccount, InitRoleBinding, InitJob},
}

var InitJob = &SubModule{
	getObj: func() Object {
		return &v1.Job{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		out := &v1.Job{}
		out.Name = fmt.Sprintf("%s-init", c.Name)
		out.Namespace = c.Namespace
		out.Labels = map[string]string{
			"cluster": c.Name,
			"app":     out.Name,
		}
		out.Spec = v1.JobSpec{
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   out.Name,
					Labels: out.Labels,
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{{
						Name:    "init",
						Image:   c.Spec.InitSpec.Image,
						Command: []string{"sh", "/home/init.sh"},
						Env: []v12.EnvVar{{
							Name:  "APISERVER_ADDRESS",
							Value: "apiserver",
						}, {
							Name:  "FRONT_APISERVER_ADDRESS",
							Value: c.Spec.AccessSpec.Address,
						}, {
							Name:  "FRONT_APISERVER_PORT",
							Value: c.Spec.AccessSpec.Port,
						}, {
							Name:  "CA_PKI_NAME",
							Value: getCAPkiName(c),
						}, {
							Name:  "ETCD_PKI_NAME",
							Value: getEtcdPkiName(c),
						}, {
							Name:  "K8S_SERVER_NAME",
							Value: getServerName(c),
						}, {
							Name:  "K8S_CLIENT_NAME",
							Value: getClientName(c),
						}, {
							Name:  "ADMIN_CONFIG_NAME",
							Value: getAdminConfigName(c),
						}, {
							Name:  "NODE_CONFIG_NAME",
							Value: getNodeConfigName(c),
						}},
					}},
					RestartPolicy:      v12.RestartPolicyNever,
					ServiceAccountName: fmt.Sprintf("%s-admin", c.Name),
				},
			},
		}
		return out
	},
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		job := object.(*v1.Job)
		c.Status.Init.Status = job.Status
		c.Status.Init.Name = job.Name
		envs := job.Spec.Template.Spec.Containers[0].Env
		for _, env := range envs {
			if env.Name == "CA_PKI_NAME" {
				c.Status.Init.CaPkiName = env.Value
				continue
			}
			if env.Name == "ETCD_PKI_NAME" {
				c.Status.Init.EtcdPkiName = env.Value
				continue
			}
			if env.Name == "K8S_SERVER_NAME" {
				c.Status.Init.ServerName = env.Value
				continue
			}
			if env.Name == "K8S_CLIENT_NAME" {
				c.Status.Init.ClientName = env.Value
				continue
			}
			if env.Name == "ADMIN_CONFIG_NAME" {
				c.Status.Init.AdminConfigName = env.Value
				continue
			}
			if env.Name == "NODE_CONFIG_NAME" {
				c.Status.Init.NodeConfigName = env.Value
				continue
			}
		}
	},
	delete: func(ctx context.Context, c *tanxv1.Cluster, client client.Client) error {
		var err error
		nameFunc := []func(cluster *tanxv1.Cluster) string{getCAPkiName, getEtcdPkiName, getServerName, getClientName, getNodeConfigName, getAdminConfigName}
		for _, namef := range nameFunc {
			err = client.Delete(ctx, &v12.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: namef(c), Namespace: c.Namespace},
			})
			if err != nil && errors.IsNotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}
		}
		return err
	},
}

func getCAPkiName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-ca-pki", c.Name)
}

func getEtcdPkiName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-pki", c.Name)
}

func getServerName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-k8s-server", c.Name)
}

func getClientName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-k8s-client", c.Name)
}

func getNodeConfigName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-node-config", c.Name)
}

func getAdminConfigName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-admin-config", c.Name)
}

var InitServiceAccount = &SubModule{
	getObj: func() Object {
		return &v12.ServiceAccount{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		out := &v12.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-admin", c.Name),
				Namespace: c.Namespace,
			},
		}
		return out
	},
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		out := object.(*v12.ServiceAccount)
		c.Status.Init.ServiceAccountName = out.Name
	},
}

var InitRoleBinding = &SubModule{
	getObj: func() Object {
		return &v13.RoleBinding{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		out := &v13.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-admin", c.Name),
				Namespace: c.Namespace,
			},
			Subjects: []v13.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      fmt.Sprintf("%s-admin", c.Name),
					Namespace: c.Namespace,
				},
			},
			RoleRef: v13.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
			},
		}
		return out
	},
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		out := object.(*v13.RoleBinding)
		c.Status.Init.RoleBindingName = out.Name
	},
}
