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

type Object interface {
	runtime.Object
	metav1.Object
	metav1.ObjectMetaAccessor
}

type Module interface {
	Init(ctx context.Context, c *v1.Cluster, r *ClusterReconciler, rl logr.Logger)
	Exist() (bool, error)
	Create() error
	StatusUpdate() error
	Delete() error
	Ready() bool
}

var modules = []ParentModule{initModule, etcdModule, apiServerModule, ctrMgtModule, schedulerModule, clientModule, installPostModule}
var VersionsModules = map[string][]ParentModule{
	"1.18.4": modules,
}

type ParentModule struct {
	Name string
	context.Context
	c *v1.Cluster
	logr.Logger
	Sub []Module
	r   *ClusterReconciler
}

func (i *ParentModule) Ready() bool {
	for _, module := range i.Sub {
		if !module.Ready() {
			return false
		}
	}
	return true
}

func (i ParentModule) copy() Module {
	return &i
}

func (i *ParentModule) Delete() error {
	for _, module := range i.Sub {
		if err := module.Delete(); err != nil {
			return err
		}
	}
	return nil
}

func (i *ParentModule) Init(ctx context.Context, c *v1.Cluster, r *ClusterReconciler, rl logr.Logger) {
	i.Context = ctx
	i.c = c
	i.Logger = rl
	i.r = r
	for _, module := range i.Sub {
		module.Init(ctx, c, r, rl)
	}
}

func (i *ParentModule) Exist() (bool, error) {
	d := true
	for _, module := range i.Sub {
		exist, err := module.Exist()
		if err != nil {
			return false, err
		}
		d = d && exist
		if !d {
			return false, nil
		}
	}
	return true, nil
}

func (i *ParentModule) Create() error {
	for _, module := range i.Sub {
		if err := module.Create(); err != nil {
			return err
		}
	}
	return nil
}

func (i *ParentModule) StatusUpdate() error {
	for _, module := range i.Sub {
		if err := module.StatusUpdate(); err != nil {
			return err
		}
	}
	return nil
}

type SubModule struct {
	context.Context
	c *v1.Cluster
	logr.Logger
	target Object
	r      *ClusterReconciler

	getObj       func() Object
	render       func(c *v1.Cluster, s *SubModule) Object
	updateStatus func(c *v1.Cluster, object Object)
	//delete not set owner obj
	delete func(ctx context.Context, c *v1.Cluster, client client.Client) error
	ready  func(c *v1.Cluster) bool
}

func (s *SubModule) Ready() bool {
	if s.ready == nil {
		return true
	}
	return s.ready(s.c)
}

func (s *SubModule) Delete() error {
	if s.delete != nil {
		err := s.delete(s, s.c, s.r.Client)
		s.Logger.Info("call subModule custom delete result", "error", err)
		return err
	}
	return nil
}

func (s *SubModule) Init(ctx context.Context, c *v1.Cluster, r *ClusterReconciler, rl logr.Logger) {
	s.Context = ctx
	s.c = c
	s.r = r
	s.Logger = rl

	s.target = s.render(s.c, s)
}

func (s *SubModule) Exist() (bool, error) {
	err := s.r.Get(s, types.NamespacedName{
		Namespace: s.c.Namespace,
		Name:      s.target.GetName(),
	}, s.getObj())
	if err != nil && errors.IsNotFound(err) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func (s *SubModule) Create() error {
	if err := controllerutil.SetControllerReference(s.c, s.target, s.r.Scheme); err != nil {
		return err
	}
	s.r.Recorder.Event(s.c, v12.EventTypeNormal, "Creating", fmt.Sprintf("%s", s.target.GetName()))
	err := s.r.Client.Create(s, s.target)
	if err != nil && errors.IsAlreadyExists(err) {
		return nil
	}
	if err != nil {
		s.r.Recorder.Event(s.c, v12.EventTypeWarning, "CreateError", fmt.Sprintf("%s,error:%v", s.target.GetName(), err))
	}
	return err
}

func (s *SubModule) StatusUpdate() error {
	out := s.getObj()
	if err := s.r.Client.Get(s, client.ObjectKey{
		Namespace: s.c.Namespace,
		Name:      s.target.GetName(),
	}, out); err != nil {
		return err
	}
	s.updateStatus(s.c, out)
	return nil
}
