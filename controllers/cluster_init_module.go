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
	"math/big"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var initModule = ParentModule{
	Name: "init-job",
	Sub:  []Module{InitServiceAccount, InitRoleBinding, InitJob},
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
							Value: fmt.Sprintf("%s-apiserver", c.Name),
						}, {
							Name:  "FRONT_APISERVER_ADDRESS",
							Value: fmt.Sprintf("%s-apiserver", c.Name),
						}, {
							Name:  "FRONT_APISERVER_PORT",
							Value: c.Spec.AccessSpec.Port,
						}, {
							Name:  "KUBE_SVC_ADDR",
							Value: NextIpForRange(c.Spec.ServiceClusterIpRange, 1),
						}, {
							Name:  "CA_PKI_NAME",
							Value: getCAPkiName(c),
						}, {
							Name:  "ETCD_SVC_NAME",
							Value: getEtcdSvcName(c),
						}, {
							Name:  "ETCD_SVC_CLIENT_NAME",
							Value: getEtcdSvcClientName(c),
						}, {
							Name:  "ETCD_PKI_PEER_NAME",
							Value: getEtcdPkiPeerName(c),
						}, {
							Name:  "ETCD_PKI_SERVER_NAME",
							Value: getEtcdPkiServerName(c),
						}, {
							Name:  "ETCD_PKI_CLIENT_NAME",
							Value: getEtcdPkiClientName(c),
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
						}, {
							Name:  "CLUSTER_DOMAIN",
							Value: c.Spec.ClusterDomain,
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
		c.Status.Init.DnsAddr = NextIpForRange(c.Spec.ServiceClusterIpRange, 2)
		envs := job.Spec.Template.Spec.Containers[0].Env
		for _, env := range envs {
			if env.Name == "CA_PKI_NAME" {
				c.Status.Init.CaPkiName = env.Value
				continue
			}
			if env.Name == "ETCD_PKI_PEER_NAME" {
				c.Status.Init.EtcdPkiPeerName = env.Value
				continue
			}
			if env.Name == "ETCD_PKI_SERVER_NAME" {
				c.Status.Init.EtcdPkiServerName = env.Value
				continue
			}
			if env.Name == "ETCD_PKI_CLIENT_NAME" {
				c.Status.Init.EtcdPkiClientName = env.Value
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
		nameFunc := []func(cluster *tanxv1.Cluster) string{getCAPkiName, getEtcdPkiClientName, getEtcdPkiServerName, getEtcdPkiPeerName, getServerName, getClientName, getNodeConfigName, getAdminConfigName}
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
	ready: func(c *tanxv1.Cluster) bool {
		for _, condition := range c.Status.Init.Status.Conditions {
			if v1.JobComplete == condition.Type {
				return true
			}
		}
		return false
	},
}

func getEtcdSvcClientName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-client", c.Name)
}

func getEtcdSvcName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd", c.Name)
}

func getEtcdPkiClientName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-pki-client", c.Name)
}

func getEtcdPkiServerName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-pki-server", c.Name)
}

func getEtcdPkiPeerName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-pki-peer", c.Name)
}

func NextIpForRange(ipRange string, step int64) string {
	ip, _, _ := net.ParseCIDR(ipRange)
	ret := big.NewInt(0)
	ret.SetBytes(ip.To4())
	i := ret.Int64() + step
	return fmt.Sprintf("%d.%d.%d.%d", byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

func getCAPkiName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-ca-pki", c.Name)
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
