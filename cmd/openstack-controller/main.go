/*
Copyright (c) 2018 OpenStack Foundation.
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
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/cloud-provider-openstack/pkg/cloud-controller/rbac"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/golang/glog"
	"golang.org/x/sync/errgroup"
)


func startControllers(kubeClient *kubernetes.Clientset, /*osClient openstack.Interface*/) error {

	// Creates a new RBAC controller
	rbacController, err := rbaccontroller.NewRBACController(kubeClient /*osClient*/)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg, ctx := errgroup.WithContext(ctx)

	// start rbac controller
	wg.Go(func() error { return rbacController.Run(ctx.Done()) })

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		glog.V(4).Info("Received SIGTERM, exiting gracefully...")
	case <-ctx.Done():
	}

	cancel()
	if err := wg.Wait(); err != nil {
		glog.Errorf("Unhandled error received: %v", err)
		return err
	}

	return nil
}

func initClients() (*kubernetes.Clientset, /*openstack.Interface,*/error) {
	// Create kubernetes client config.
	config, err := newClusterConfig("/etc/kubernetes/admin.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	// TODO: Create OpenStack client from config file.

	return kubeClient, /*osClient,*/ nil
}


// NewClusterConfig builds a kubernetes cluster config.
func newClusterConfig(kubeConfig string) (*rest.Config, error) {
	var cfg *rest.Config
	var err error

	if kubeConfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func main() {
	// Initilize kubernetes and openstack clients.
	kubeClient, /*osClient,*/ err := initClients()
	if err != nil {
		glog.Fatal(err)
	}

	// Start controllers.
	if err := startControllers(kubeClient, /*osClient*/); err != nil {
		glog.Fatal(err)
	}
}
