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
	"k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateRoleByNamespace generates default-role which has all the permissions in the namespace.
func GenerateRoleByNamespace(namespace string) *v1.Role {
	policyRule := v1.PolicyRule{
		Verbs:     []string{v1.VerbAll},
		APIGroups: []string{v1.APIGroupAll},
		Resources: []string{v1.ResourceAll},
	}
	role := &v1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-role",
			Namespace: namespace,
		},
		Rules: []v1.PolicyRule{policyRule},
	}
	return role
}

// GenerateRoleBinding generates rolebinding which allows user has default-role in the project namespace.
func GenerateRoleBinding(namespace, project string) *v1.RoleBinding {
	subject := v1.Subject{
		Kind: "User",
		Name: project,
	}
	roleRef := v1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "default-role",
	}
	roleBinding := &v1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      project + "-rolebinding",
			Namespace: namespace,
		},
		Subjects: []v1.Subject{subject},
		RoleRef:  roleRef,
	}
	return roleBinding
}

// GenerateServiceAccountRoleBinding generates rolebinding of service account in the namespace.
func GenerateServiceAccountRoleBinding(namespace, project string) *v1.RoleBinding {
	subject := v1.Subject{
		Kind:      "ServiceAccount",
		Name:      "default",
		Namespace: namespace,
	}
	roleRef := v1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "default-role",
	}
	roleBinding := &v1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      project + "-rolebinding-sa",
			Namespace: namespace,
		},
		Subjects: []v1.Subject{subject},
		RoleRef:  roleRef,
	}
	return roleBinding
}

// GenerateClusterRole generates namespace-creater ClusterRole which has the permission of namespaces resource.
func GenerateClusterRole() *v1.ClusterRole {
	policyRule := v1.PolicyRule{
		Verbs:     []string{v1.VerbAll},
		APIGroups: []string{v1.APIGroupAll},
		Resources: []string{"namespaces"},
	}

	clusterRole := &v1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace-creater",
		},
		Rules: []v1.PolicyRule{policyRule},
	}
	return clusterRole
}

// GenerateClusterRoleBindingByProject generates ClusterRoleBinding which allows anyone in the "project" group to create namespace.
func GenerateClusterRoleBindingByProject(project string) *v1.ClusterRoleBinding {
	subject := v1.Subject{
		Kind: "Group",
		Name: project,
	}
	roleRef := v1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "namespace-creater",
	}

	clusterRoleBinding := &v1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: project + "-namespace-creater",
		},
		Subjects: []v1.Subject{subject},
		RoleRef:  roleRef,
	}
	return clusterRoleBinding
}
