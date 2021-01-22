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

package main

import (
	"flag"
	"fmt"
	vs1 "github.com/tangxusc/kok/controllers/cluster1"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	clusterv1 "github.com/tangxusc/kok/api/v1"
	"github.com/tangxusc/kok/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = clusterv1.AddToScheme(scheme)
	_ = clusterv1.EtcdAddToScheme(scheme)
	_ = vs1.Version
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	fmt.Println("目前支持的版本情况:")
	for v, modules := range controllers.VersionsModules {
		names := make([]string, len(modules))
		for i, module := range modules {
			names[i] = module.Name
		}
		fmt.Printf("版本:%s,模块[%s]\n", v, strings.Join(names, ","))
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "9c2a81bf.kok.tanx",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ClusterReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("Cluster"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("Cluster"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}
	if err = (&clusterv1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Cluster")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	go startAddonsDownloader(mgr)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

const addonsDirName = "addons"

var version2Addons = map[string]map[string]*template.Template{}
var defaultTemplateFuncs = template.FuncMap{"join": strings.Join}

func startAddonsDownloader(mgr manager.Manager) {
	c := mgr.GetClient()

	err := initTemplateMaps()
	if err != nil {
		setupLog.Error(err, "load addons template error")
	}
	engine := gin.Default()
	engine.Any("/download/:namespace/:name/:dir/:filename", func(ctx *gin.Context) {
		ns := ctx.Param("namespace")
		name := ctx.Param("name")
		dir := ctx.Param("dir")
		filename := ctx.Param("filename")

		cls := &clusterv1.Cluster{}
		err := c.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, cls)
		if err != nil {
			panic(err)
		}

		t, ok := version2Addons[cls.Spec.ClusterVersion][dir]
		if !ok {
			if _, err := ctx.Writer.WriteString("未找到文件模板,传入的dir可能存在错误"); err != nil {
				panic(err)
			}
		}
		if err = t.ExecuteTemplate(ctx.Writer, filename, cls); err != nil {
			panic(err)
		}
		ctx.Writer.Flush()
	})
	engine.Any("/meta/:namespace/:name/ca/:filename", func(ctx *gin.Context) {
		getMeta(ctx, c, "ca")
	})

	engine.Any("/meta/:namespace/:name/nodeconfig/:filename", func(ctx *gin.Context) {
		getMeta(ctx, c, "nodeconfig")
	})

	if err := engine.Run(":7788"); err != nil {
		setupLog.Error(err, "start addons downloader error")
	}
}

func getMeta(ctx *gin.Context, client client.Client, dir string) {
	ns := ctx.Param("namespace")
	name := ctx.Param("name")
	filename := ctx.Param("filename")

	cls := &clusterv1.Cluster{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, cls)
	if err != nil {
		panic(err)
	}

	sourceName := ""
	switch dir {
	case "ca":
		sourceName = cls.Status.Init.CaPkiName
	case "nodeconfig":
		sourceName = cls.Status.Init.NodeConfigName
	}
	ca := &v1.Secret{}
	err = client.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      sourceName,
	}, ca)
	if err != nil {
		panic(err)
	}
	b, ok := ca.Data[filename]
	if !ok {
		if _, err := ctx.Writer.WriteString("未找到元数据,传入的filename可能存在错误"); err != nil {
			panic(err)
		}
	}
	fmt.Printf("%s", b)

	if _, err := ctx.Writer.Write(b); err != nil {
		panic(err)
	}
	ctx.Writer.Flush()
}

func initTemplateMaps() error {
	dir, err := ioutil.ReadDir(addonsDirName)
	if err != nil {
		return err
	}
	for _, sub := range dir {
		if !sub.IsDir() {
			continue
		}
		join := filepath.Join(addonsDirName, sub.Name())
		subDir, err := ioutil.ReadDir(join)
		if err != nil {
			return err
		}
		m := map[string]*template.Template{}
		for _, info := range subDir {
			if !info.IsDir() {
				continue
			}
			name := filepath.Join(join, info.Name())
			t, err := template.New(name).Funcs(defaultTemplateFuncs).ParseGlob(name + "/*")
			if err != nil {
				return err
			}
			m[info.Name()] = t
		}

		version2Addons[sub.Name()] = m
	}
	return nil
}
