/*

 Copyright 2019 The KubeSphere Authors.

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
package app

import (
	goflag "flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/options"
	"kubesphere.io/kubesphere/pkg/signals"
	"log"
	"net/http"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.SharedOptions

	cmd := &cobra.Command{
		Use: "ks-apiserver",
		Long: `The KubeSphere API server validates and configures data
for the api objects. The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(s)
		},
	}

	cmd.Flags().AddFlagSet(s.CommandLine)
	cmd.Flags().AddGoFlagSet(goflag.CommandLine)
	glog.CopyStandardLogTo("INFO")
	return cmd
}

func Run(s *options.ServerRunOptions) error {

	var err error

	waitForResourceSync()

	container := runtime.Container
	container.Filter(filter.Logging)

	log.Printf("Server listening on %d.", s.InsecurePort)

	if s.InsecurePort != 0 {
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", s.BindAddress, s.InsecurePort), container)
	}

	if s.SecurePort != 0 && len(s.TlsCertFile) > 0 && len(s.TlsPrivateKey) > 0 {
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.BindAddress, s.SecurePort), s.TlsCertFile, s.TlsPrivateKey, container)
	}

	return err
}

func waitForResourceSync() {
	stopChan := signals.SetupSignalHandler()

	informerFactory := informers.SharedInformerFactory()
	informerFactory.Rbac().V1().Roles().Lister()
	informerFactory.Rbac().V1().RoleBindings().Lister()
	informerFactory.Rbac().V1().ClusterRoles().Lister()
	informerFactory.Rbac().V1().ClusterRoleBindings().Lister()

	informerFactory.Storage().V1().StorageClasses().Lister()

	informerFactory.Core().V1().Namespaces().Lister()
	informerFactory.Core().V1().Nodes().Lister()
	informerFactory.Core().V1().ResourceQuotas().Lister()
	informerFactory.Core().V1().Pods().Lister()
	informerFactory.Core().V1().Services().Lister()
	informerFactory.Core().V1().PersistentVolumeClaims().Lister()
	informerFactory.Core().V1().Secrets().Lister()
	informerFactory.Core().V1().ConfigMaps().Lister()

	informerFactory.Apps().V1().ControllerRevisions().Lister()
	informerFactory.Apps().V1().StatefulSets().Lister()
	informerFactory.Apps().V1().Deployments().Lister()
	informerFactory.Apps().V1().DaemonSets().Lister()

	informerFactory.Batch().V1().Jobs().Lister()
	informerFactory.Batch().V1beta1().CronJobs()

	informerFactory.Start(stopChan)
	informerFactory.WaitForCacheSync(stopChan)
	log.Println("resources sync success")
}