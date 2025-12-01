/*
Copyright The Kubernetes Authors.

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

package daemonset

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/kueue/pkg/controller/jobframework"
)

var (
	gvk = appsv1.SchemeGroupVersion.WithKind("DaemonSet")
)

const (
	FrameworkName = "daemonset"
)

func init() {
	utilruntime.Must(jobframework.RegisterIntegration(FrameworkName, jobframework.IntegrationCallbacks{
		SetupIndexes:           SetupIndexes,
		NewReconciler:          jobframework.NewNoopReconcilerFactory(gvk),
		GVK:                    gvk,
		SetupWebhook:           SetupWebhook,
		JobType:                &appsv1.DaemonSet{},
		AddToScheme:            appsv1.AddToScheme,
		DependencyList:         []string{"pod"},
		IsManagingObjectsOwner: isDaemonSet,
	}))
}

func isDaemonSet(ref *metav1.OwnerReference) bool {
	return ref.Kind == "DaemonSet" && ref.APIVersion == "apps/v1"
}

type DaemonSet appsv1.DaemonSet

// +kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=get;list;watch

func fromObject(o runtime.Object) *DaemonSet {
	return (*DaemonSet)(o.(*appsv1.DaemonSet))
}

func (d *DaemonSet) Object() client.Object {
	return (*appsv1.DaemonSet)(d)
}

func (d *DaemonSet) GVK() schema.GroupVersionKind {
	return gvk
}

func SetupIndexes(context.Context, client.FieldIndexer) error {
	return nil
}