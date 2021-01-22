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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const FinalizerName = "finalizer.cluster.kok.tanx"

var VersionsModules = make(map[string][]*Module)

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

	for i := 0; i < len(value); i++ {
		for j := 0; j < len(value); j++ {
			if value[i].Order < value[j].Order {
				value[i], value[j] = value[j], value[i]
			}
		}
	}

	VersionsModules[key] = value
	for _, module := range m {
		v1.RegisterVersionedDefaulters(key, module)
		v1.RegisterVersionedValidators(key, module)
	}
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
	Name  string
	Sub   []*Module
	Order int

	GetObj               func() Object
	Render               func(c *v1.Cluster) Object
	SetStatus            func(c *v1.Cluster, target, now Object) (bool, Object)
	Del                  func(ctx context.Context, c *v1.Cluster, client client.Client) error
	Next                 func(c *v1.Cluster) bool
	SetDefault           func(c *v1.Cluster)
	ValidateCreateModule func(c *v1.Cluster) field.ErrorList
	ValidateUpdateModule func(now *v1.Cluster, old *v1.Cluster) field.ErrorList
}

func (m *Module) ValidateCreate(c *v1.Cluster) field.ErrorList {
	if !m.hasSub() {
		if m.ValidateCreateModule == nil {
			return nil
		}
		return m.ValidateCreateModule(c)
	} else {
		for _, m := range m.Sub {
			if err := m.ValidateCreate(c); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Module) ValidateUpdate(now *v1.Cluster, old *v1.Cluster) field.ErrorList {
	if !m.hasSub() {
		if m.ValidateUpdateModule == nil {
			return nil
		}
		return m.ValidateUpdateModule(now, old)
	} else {
		for _, m := range m.Sub {
			if err := m.ValidateUpdate(now, old); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Module) Default(c *v1.Cluster) {
	if !m.hasSub() {
		if m.SetDefault != nil {
			m.SetDefault(c)
		}
	} else {
		for _, m := range m.Sub {
			m.Default(c)
		}
	}
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
	render := m.Render(ctx.Cluster)
	err := ctx.Get(ctx, types.NamespacedName{
		Namespace: ctx.Namespace,
		Name:      render.GetName(),
	}, m.GetObj())
	if err != nil && errors.IsNotFound(err) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func (m *Module) update(ctx *ModuleContext) error {
	obj := m.GetObj()
	render := m.Render(ctx.Cluster)
	if err := ctx.Client.Get(ctx, client.ObjectKey{
		Namespace: ctx.Namespace,
		Name:      render.GetName(),
	}, obj); err != nil {
		return err
	}

	needUpdate, o := m.SetStatus(ctx.Cluster, render, obj)
	if needUpdate {
		if err := ctx.Client.Update(ctx.Context, o); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) create(ctx *ModuleContext) error {
	render := m.Render(ctx.Cluster)
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
		if m.Next != nil {
			return m.Next(ctx.Cluster)
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
		if m.Del != nil {
			err := m.Del(ctx, ctx.Cluster, ctx.Client)
			if err != nil {
				ctx.Logger.Error(err, "invoke module Del error")
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
