package controllers

import (
	"context"
	"github.com/go-logr/logr"
	v1 "github.com/tangxusc/kok/api/v1"
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
	Init(ctx context.Context, c *v1.Cluster, r client.Client, rl logr.Logger, scheme *runtime.Scheme)
	Exist() (bool, error)
	Create() error
	StatusUpdate() error
	Delete() error
}

var modules = []ParentModule{initModule, etcdModule, apiServerModule, ctrMgtModule, schedulerModule, clientModule, installPostModule}
var VersionsModules = map[string][]ParentModule{
	"1.18.4": modules,
}

type ParentModule struct {
	context.Context
	c *v1.Cluster
	client.Client
	logr.Logger
	Sub []Module
	*runtime.Scheme
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

func (i *ParentModule) Init(ctx context.Context, c *v1.Cluster, r client.Client, rl logr.Logger, scheme *runtime.Scheme) {
	i.Context = ctx
	i.c = c
	i.Client = r
	i.Logger = rl
	i.Scheme = scheme
	for _, module := range i.Sub {
		module.Init(ctx, c, r, rl, scheme)
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
	client.Client
	logr.Logger
	*runtime.Scheme
	target Object

	getObj       func() Object
	render       func(c *v1.Cluster, s *SubModule) Object
	updateStatus func(c *v1.Cluster, object Object)
	//delete not set owner obj
	delete func(ctx context.Context, c *v1.Cluster, client client.Client) error
}

func (s *SubModule) Delete() error {
	if s.delete != nil {
		err := s.delete(s, s.c, s.Client)
		s.Logger.Info("call subModule custom delete result", "error", err)
		return err
	}
	return nil
}

func (s *SubModule) Init(ctx context.Context, c *v1.Cluster, r client.Client, rl logr.Logger, scheme *runtime.Scheme) {
	s.Context = ctx
	s.c = c
	s.Client = r
	s.Logger = rl
	s.Scheme = scheme

	s.target = s.render(s.c, s)
}

func (s *SubModule) Exist() (bool, error) {
	err := s.Get(s, types.NamespacedName{
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
	if err := controllerutil.SetControllerReference(s.c, s.target, s.Scheme); err != nil {
		return err
	}
	err := s.Client.Create(s, s.target)
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (s *SubModule) StatusUpdate() error {
	out := s.getObj()
	if err := s.Client.Get(s, client.ObjectKey{
		Namespace: s.c.Namespace,
		Name:      s.target.GetName(),
	}, out); err != nil {
		return err
	}
	s.updateStatus(s.c, out)
	return nil
}
