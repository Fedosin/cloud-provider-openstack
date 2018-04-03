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

package rbaccontroller

import (
	"time"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	resyncPeriod = 5 * time.Minute
)

// Controller manages life cycle of namespace's rbac.
type Controller struct {
	k8sclient     kubernetes.Interface
}

// NewRBACController creates a new RBAC controller.
func NewRBACController(kubeClient kubernetes.Interface) (*Controller, error) {
	c := &Controller{
		k8sclient:     kubeClient,
	}

	return c, nil
}

// Run the controller.
func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	source := cache.NewListWatchFromClient(
		c.k8sclient.Core().RESTClient(),
		"namespaces",
		apiv1.NamespaceAll,
		fields.Everything())

	_, namespaceInformor := cache.NewInformer(
		source,
		&apiv1.Namespace{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		})

	go namespaceInformor.Run(stopCh)
	<-stopCh
	return nil
}

func (c *Controller) onAdd(obj interface{}) {
	namespace := obj.(*apiv1.Namespace)
	glog.V(3).Infof("RBAC controller received new object %#v\n", namespace)

	c.syncRBAC(namespace)
}

func (c *Controller) onUpdate(obj1, obj2 interface{}) {
	namespace := obj1.(*apiv1.Namespace)
	glog.V(3).Infof("RBAC controller received changed object %#v\n", namespace)

	c.syncRBAC(namespace)
}

func (c *Controller) onDelete(obj interface{}) {
	namespace := obj.(*apiv1.Namespace)
	// rbac controller have done all the works so we will not wait here
	glog.V(3).Infof("RBAC controller received deleted namespace %#v\n", namespace)
}

func (c *Controller) syncRBAC(ns *apiv1.Namespace) error {
	if ns.DeletionTimestamp != nil {
		return nil
	}
	rbacClient := c.k8sclient.Rbac()

	// Create role for project
	role := GenerateRoleByNamespace(ns.Name)
	_, err := rbacClient.Roles(ns.Name).Create(role)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		glog.Errorf("Failed create default-role in namespace %s for project %s: %v", ns.Name, ns.Name, err)
		return err
	}
	glog.V(4).Infof("Created default-role in namespace %s for project %s", ns.Name, ns.Name)

	// Create rolebinding for project
	roleBinding := GenerateRoleBinding(ns.Name, ns.Name)
	_, err = rbacClient.RoleBindings(ns.Name).Create(roleBinding)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		glog.Errorf("Failed create %s-rolebindings in namespace %s for project %s: %v", ns.Name, ns.Name, ns.Name, err)
		return err
	}
	saRoleBinding := GenerateServiceAccountRoleBinding(ns.Name, ns.Name)
	_, err = rbacClient.RoleBindings(ns.Name).Create(saRoleBinding)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		glog.Errorf("Failed create %s-rolebindings-sa in namespace %s for project %s: %v", ns.Name, ns.Name, ns.Name, err)
		return err
	}

	glog.V(4).Infof("Created %s-rolebindings in namespace %s for project %s", ns.Name, ns.Name, ns.Name)
	return nil
}
