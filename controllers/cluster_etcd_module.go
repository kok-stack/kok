package controllers

import (
	"fmt"
	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var etcdModule = ParentModule{
	Name: "etcd-crd",
	Sub:  []Module{etcdCRD},
}

var etcdCRD = &SubModule{
	getObj: func() Object {
		return &v1beta2.EtcdCluster{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		name := fmt.Sprintf("%s-etcd", c.Name)
		var out = &v1beta2.EtcdCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: c.Namespace,
			},
			Spec: v1beta2.ClusterSpec{
				Size: c.Spec.EtcdSpec.Count,
				TLS: &v1beta2.TLSPolicy{
					Static: &v1beta2.StaticTLS{
						Member: &v1beta2.MemberSecret{
							PeerSecret:   c.Status.Init.EtcdPkiPeerName,
							ServerSecret: c.Status.Init.EtcdPkiServerName,
						},
						OperatorSecret: c.Status.Init.EtcdPkiClientName,
					},
				},
			},
		}
		return out
	},
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		obj := object.(*v1beta2.EtcdCluster)
		c.Status.Etcd.SvcName = obj.Status.ServiceName
		c.Status.Etcd.Name = obj.Name
		c.Status.Etcd.Status = obj.Status
	},
	ready: func(c *tanxv1.Cluster) bool {
		if len(c.Status.Etcd.Status.Members.Ready) == c.Status.Etcd.Status.Size && (c.Status.Etcd.Status.Size == c.Spec.EtcdSpec.Count) {
			return true
		}
		return false
	},
}
