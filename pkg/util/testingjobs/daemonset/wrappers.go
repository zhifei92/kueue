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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/kueue/pkg/constants"
	controllerconstants "sigs.k8s.io/kueue/pkg/controller/constants"
)

// DaemonSetWrapper wraps a DaemonSet.
type DaemonSetWrapper struct {
	appsv1.DaemonSet
}

// MakeDaemonSet creates a wrapper for a DaemonSet with a single container.
func MakeDaemonSet(name, ns string) *DaemonSetWrapper {
	podLabels := map[string]string{
		"app": fmt.Sprintf("%s-pod", name),
	}
	return &DaemonSetWrapper{appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: make(map[string]string, 1),
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "c",
							Image:     "pause",
							Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{}},
						},
					},
					NodeSelector: map[string]string{},
				},
			},
		},
	}}
}

// Obj returns the inner DaemonSet.
func (d *DaemonSetWrapper) Obj() *appsv1.DaemonSet {
	return &d.DaemonSet
}

// Label sets the label of the DaemonSet
func (d *DaemonSetWrapper) Label(k, v string) *DaemonSetWrapper {
	if d.Labels == nil {
		d.Labels = make(map[string]string)
	}
	d.Labels[k] = v
	return d
}

// Queue updates the queue name of the DaemonSet
func (d *DaemonSetWrapper) Queue(q string) *DaemonSetWrapper {
	return d.Label(controllerconstants.QueueLabel, q)
}

// Name updated the name of the DaemonSet
func (d *DaemonSetWrapper) Name(n string) *DaemonSetWrapper {
	d.ObjectMeta.Name = n
	return d
}

// UID updates the uid of the DaemonSet.
func (d *DaemonSetWrapper) UID(uid string) *DaemonSetWrapper {
	d.ObjectMeta.UID = types.UID(uid)
	return d
}

// Image sets an image to the default container.
func (d *DaemonSetWrapper) Image(image string, args []string) *DaemonSetWrapper {
	d.Spec.Template.Spec.Containers[0].Image = image
	d.Spec.Template.Spec.Containers[0].Args = args
	return d
}

// Request adds a resource request to the default container.
func (d *DaemonSetWrapper) Request(r corev1.ResourceName, v string) *DaemonSetWrapper {
	if d.Spec.Template.Spec.Containers[0].Resources.Requests == nil {
		d.Spec.Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{}
	}
	d.Spec.Template.Spec.Containers[0].Resources.Requests[r] = resource.MustParse(v)
	return d
}

// Limit adds a resource limit to the default container.
func (d *DaemonSetWrapper) Limit(r corev1.ResourceName, v string) *DaemonSetWrapper {
	if d.Spec.Template.Spec.Containers[0].Resources.Limits == nil {
		d.Spec.Template.Spec.Containers[0].Resources.Limits = corev1.ResourceList{}
	}
	d.Spec.Template.Spec.Containers[0].Resources.Limits[r] = resource.MustParse(v)
	return d
}

// RequestAndLimit adds a resource request and limit to the default container.
func (d *DaemonSetWrapper) RequestAndLimit(r corev1.ResourceName, v string) *DaemonSetWrapper {
	return d.Request(r, v).Limit(r, v)
}

// NumberReady updates the numberReady of the DaemonSet
func (d *DaemonSetWrapper) NumberReady(numberReady int32) *DaemonSetWrapper {
	d.Status.NumberReady = numberReady
	return d
}

// PodTemplateSpecLabel sets the label of the pod template spec of the DaemonSet
func (d *DaemonSetWrapper) PodTemplateSpecLabel(k, v string) *DaemonSetWrapper {
	if d.Spec.Template.Labels == nil {
		d.Spec.Template.Labels = make(map[string]string, 1)
	}
	d.Spec.Template.Labels[k] = v
	return d
}

// PodTemplateAnnotation sets the annotation of the pod template
func (d *DaemonSetWrapper) PodTemplateAnnotation(k, v string) *DaemonSetWrapper {
	if d.Spec.Template.Annotations == nil {
		d.Spec.Template.Annotations = make(map[string]string, 1)
	}
	d.Spec.Template.Annotations[k] = v
	return d
}

// PodTemplateSpecQueue updates the queue name of the pod template spec of the DaemonSet
func (d *DaemonSetWrapper) PodTemplateSpecQueue(q string) *DaemonSetWrapper {
	return d.PodTemplateSpecLabel(controllerconstants.QueueLabel, q)
}

func (d *DaemonSetWrapper) PodTemplateSpecManagedByKueue() *DaemonSetWrapper {
	return d.PodTemplateSpecLabel(constants.ManagedByKueueLabelKey, constants.ManagedByKueueLabelValue)
}

func (d *DaemonSetWrapper) TerminationGracePeriod(seconds int64) *DaemonSetWrapper {
	d.Spec.Template.Spec.TerminationGracePeriodSeconds = &seconds
	return d
}

func (d *DaemonSetWrapper) SetTypeMeta() *DaemonSetWrapper {
	d.APIVersion = appsv1.SchemeGroupVersion.String()
	d.Kind = "DaemonSet"
	return d
}