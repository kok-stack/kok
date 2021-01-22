/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strconv"
	"strings"
)

// log is for logging in this package.
var clusterlog = logf.Log.WithName("cluster-resource")

func (r *Cluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-cluster-kok-tanx-v1-cluster,mutating=true,failurePolicy=fail,groups=cluster.kok.tanx,resources=clusters,verbs=create;update,versions=v1,name=mcluster.kb.io

var _ webhook.Defaulter = &Cluster{}

//+kubebuilder:object:generate:=false
//+kubebuilder:object:root:=false
type ClusterValidator interface {
	ValidateCreate(c *Cluster) field.ErrorList
	ValidateUpdate(now *Cluster, old *Cluster) field.ErrorList
}

//+kubebuilder:object:generate:=false
//+kubebuilder:object:root:=false
type ClusterDefaulter interface {
	Default(c *Cluster)
}

var VersionedValidators = map[string][]ClusterValidator{}
var VersionedDefaulters = map[string][]ClusterDefaulter{}

func RegisterVersionedValidators(version string, o ClusterValidator) {
	setMaxVersion(version)
	value, ok := VersionedValidators[version]
	if !ok {
		value = make([]ClusterValidator, 0)
	}
	value = append(value, o)
	VersionedValidators[version] = value
}

func setMaxVersion(version string) {
	if maxVersion == "" {
		maxVersion = version
		return
	}
	now, err := strconv.Atoi(strings.ReplaceAll(version, ".", ""))
	if err != nil {
		return
	}
	old, err := strconv.Atoi(strings.ReplaceAll(maxVersion, ".", ""))
	if err != nil {
		return
	}
	if now > old {
		maxVersion = version
	}
	fmt.Println("maxVersion:", maxVersion)
}

func RegisterVersionedDefaulters(version string, o ClusterDefaulter) {
	value, ok := VersionedDefaulters[version]
	if !ok {
		value = make([]ClusterDefaulter, 0)
	}
	value = append(value, o)
	VersionedDefaulters[version] = value
}

var maxVersion = ""

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cluster) Default() {
	clusterlog.Info("default", "name", r.Name)

	v := getVersion(r.Spec.ClusterVersion)
	defaulters := VersionedDefaulters[v]
	for _, defaulter := range defaulters {
		defaulter.Default(r)
	}
}

func getVersion(version string) string {
	if version != "" {
		return version
	}
	return maxVersion
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cluster-kok-tanx-v1-cluster,mutating=false,failurePolicy=fail,groups=cluster.kok.tanx,resources=clusters,versions=v1,name=vcluster.kb.io

var _ webhook.Validator = &Cluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateCreate() error {
	clusterlog.Info("validate create", "name", r.Name)
	var allErrs field.ErrorList

	validators := VersionedValidators[r.Spec.ClusterVersion]
	for _, v := range validators {

		allErrs = append(allErrs, v.ValidateCreate(r)...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateUpdate(old runtime.Object) error {
	clusterlog.Info("validate update", "name", r.Name)
	oldC := old.(*Cluster)
	var allErrs field.ErrorList

	validators := VersionedValidators[r.Spec.ClusterVersion]
	for _, v := range validators {

		allErrs = append(allErrs, v.ValidateUpdate(oldC, r)...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateDelete() error {
	clusterlog.Info("validate delete", "name", r.Name)

	return nil
}
