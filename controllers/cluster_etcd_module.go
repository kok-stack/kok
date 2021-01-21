package controllers

import (
	"fmt"
	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

var etcdModule = &Module{
	Name: "etcd-crd",
	Sub:  []*Module{etcdCRD},
}

var etcdCRD = &Module{
	getObj: func() Object {
		return &v1beta2.EtcdCluster{}
	},
	render: func(c *tanxv1.Cluster) Object {
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
	setStatus: func(c *tanxv1.Cluster, target, now Object) (bool, Object) {
		obj := now.(*v1beta2.EtcdCluster)
		c.Status.Etcd.SvcName = obj.Status.ServiceName
		c.Status.Etcd.Name = obj.Name
		c.Status.Etcd.Status = obj.Status

		t := target.(*v1beta2.EtcdCluster)
		n := now.(*v1beta2.EtcdCluster)
		if !reflect.DeepEqual(t.Spec, n.Spec) {
			n.Spec = t.Spec
			return true, n
		}
		return false, n
	},
	ready: func(c *tanxv1.Cluster) bool {
		if len(c.Status.Etcd.Status.Members.Ready) == c.Status.Etcd.Status.Size && (c.Status.Etcd.Status.Size == c.Spec.EtcdSpec.Count) {
			return true
		}
		return false
	},
}
