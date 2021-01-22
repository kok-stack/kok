package cluster1

import (
	"fmt"
	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	"github.com/tangxusc/kok/controllers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
)

func init() {
	controllers.AddModules(Version, etcdModule)
}

var etcdModule = &controllers.Module{
	Order: 20,
	Name:  "etcd-crd",
	Sub:   []*controllers.Module{etcdCRD},
}

var etcdCRD = &controllers.Module{
	GetObj: func() controllers.Object {
		return &v1beta2.EtcdCluster{}
	},
	Render: func(c *tanxv1.Cluster) controllers.Object {
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
	SetStatus: func(c *tanxv1.Cluster, target, now controllers.Object) (bool, controllers.Object) {
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
	Next: func(c *tanxv1.Cluster) bool {
		if len(c.Status.Etcd.Status.Members.Ready) == c.Status.Etcd.Status.Size && (c.Status.Etcd.Status.Size == c.Spec.EtcdSpec.Count) {
			return true
		}
		return false
	},
	SetDefault: func(r *tanxv1.Cluster) {
		if r.Spec.EtcdSpec.Count == 0 {
			r.Spec.EtcdSpec.Count = 3
		}
	},
	ValidateCreateModule: func(r *tanxv1.Cluster) field.ErrorList {
		var allErrs field.ErrorList
		if r.Spec.EtcdSpec.Count%2 == 0 || r.Spec.EtcdSpec.Count < 3 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.etcdSpec.count"), r.Spec.EtcdSpec.Count, "不能为奇数且必须>=3"))
		}
		return allErrs
	},
	ValidateUpdateModule: func(now *tanxv1.Cluster, old *tanxv1.Cluster) field.ErrorList {
		var allErrs field.ErrorList
		if now.Spec.EtcdSpec.Count%2 == 0 || old.Spec.EtcdSpec.Count < 3 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.etcdSpec.count"), now.Spec.EtcdSpec.Count, "不能为奇数且必须>=3"))
		}
		return allErrs
	},
}
