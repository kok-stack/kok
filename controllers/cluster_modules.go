package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	v1 "github.com/tangxusc/kok/api/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const FinalizerName = "finalizer.cluster.kok.tanx"

var modules = []*Module{initModule, etcdModule, apiServerModule, ctrMgtModule, schedulerModule, clientModule}
var VersionsModules = map[string][]*Module{
	"1.18.4": modules,
}

type Object interface {
	runtime.Object
	metav1.Object
	metav1.ObjectMetaAccessor
}

func AddModules(key string, m ...*Module) {
	value, ok := VersionsModules[key]
	if !ok {
		value = make([]*Module, 0)
	}
	value = append(value, m...)
	VersionsModules[key] = value
}

type ModuleContext struct {
	context.Context
	*v1.Cluster
	logr.Logger
	*ClusterReconciler
}

func NewModuleContext(context context.Context, c *v1.Cluster, logger logr.Logger, r *ClusterReconciler) *ModuleContext {
	return &ModuleContext{Context: context, Cluster: c, Logger: logger, ClusterReconciler: r}
}

type Module struct {
	Name string
	Sub  []*Module

	getObj    func() Object
	render    func(c *v1.Cluster) Object
	setStatus func(c *v1.Cluster, target, now Object) (bool, Object)
	delete    func(ctx context.Context, c *v1.Cluster, client client.Client) error
	ready     func(c *v1.Cluster) bool
}

func (m *Module) Reconcile(ctx *ModuleContext) error {
	if !m.hasSub() {
		exist, err := m.exist(ctx)
		if err != nil {
			return err
		}
		if exist {
			if err := m.update(ctx); err != nil {
				return err
			}
		} else {
			if err := m.create(ctx); err != nil {
				return err
			}
		}
	} else {
		for _, m := range m.Sub {
			if err := m.Reconcile(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Module) hasSub() bool {
	if m.Sub == nil || len(m.Sub) == 0 {
		return false
	}
	return true
}

func (m *Module) exist(ctx *ModuleContext) (bool, error) {
	render := m.render(ctx.Cluster)
	err := ctx.Get(ctx, types.NamespacedName{
		Namespace: ctx.Namespace,
		Name:      render.GetName(),
	}, m.getObj())
	if err != nil && errors.IsNotFound(err) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func (m *Module) update(ctx *ModuleContext) error {
	obj := m.getObj()
	render := m.render(ctx.Cluster)
	if err := ctx.Client.Get(ctx, client.ObjectKey{
		Namespace: ctx.Namespace,
		Name:      render.GetName(),
	}, obj); err != nil {
		return err
	}

	needUpdate, o := m.setStatus(ctx.Cluster, render, obj)
	if needUpdate {
		if err := ctx.Client.Update(ctx.Context, o); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) create(ctx *ModuleContext) error {
	render := m.render(ctx.Cluster)
	if err := controllerutil.SetControllerReference(ctx.Cluster, render, ctx.Scheme); err != nil {
		return err
	}
	ctx.Recorder.Event(ctx, v12.EventTypeNormal, "Creating", render.GetName())
	err := ctx.Client.Create(ctx, render)
	if err != nil && errors.IsAlreadyExists(err) {
		return nil
	}
	if err != nil {
		ctx.Recorder.Event(ctx, v12.EventTypeWarning, "CreateError", fmt.Sprintf("%s,error:%v", render.GetName(), err))
	}
	return err
}

func (m *Module) Ready(ctx *ModuleContext) bool {
	if !m.hasSub() {
		if m.ready != nil {
			return m.ready(ctx.Cluster)
		}
		return true
	} else {
		for _, m := range m.Sub {
			if !m.Ready(ctx) {
				return false
			}
		}
	}
	return true
}

func (m *Module) Delete(ctx *ModuleContext) error {
	if !m.hasSub() {
		if m.delete != nil {
			err := m.delete(ctx, ctx.Cluster, ctx.Client)
			if err != nil {
				ctx.Logger.Error(err, "invoke module delete error")
			}
			return err
		}
	}
	for _, m := range m.Sub {
		if err := m.Delete(ctx); err != nil {
			return err
		}
	}
	return nil
}
