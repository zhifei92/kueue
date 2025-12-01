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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"sigs.k8s.io/kueue/pkg/cache"
	"sigs.k8s.io/kueue/pkg/controller/constants"
	"sigs.k8s.io/kueue/pkg/controller/jobframework"
	podconstants "sigs.k8s.io/kueue/pkg/controller/jobs/pod/constants"
	"sigs.k8s.io/kueue/pkg/features"
	"sigs.k8s.io/kueue/pkg/queue"
	utiltesting "sigs.k8s.io/kueue/pkg/util/testing"
	testingdaemonset "sigs.k8s.io/kueue/pkg/util/testingjobs/daemonset"
)

func TestDefault(t *testing.T) {
	testCases := map[string]struct {
		daemonset            *appsv1.DaemonSet
		localQueueDefaulting bool
		defaultLqExist       bool
		want                 *appsv1.DaemonSet
	}{
		"daemonset without queue": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").Obj(),
			want:      testingdaemonset.MakeDaemonSet("test-pod", "").Obj(),
		},
		"daemonset with queue": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Obj(),
			want: testingdaemonset.MakeDaemonSet("test-pod", "").
				PodTemplateSpecManagedByKueue().
				Queue("test-queue").
				PodTemplateSpecQueue("test-queue").
				PodTemplateAnnotation(podconstants.SuspendedByParentAnnotation, FrameworkName).
				Obj(),
		},
		"daemonset with queue and pod template spec queue": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("new-test-queue").
				PodTemplateSpecQueue("test-queue").
				Obj(),
			want: testingdaemonset.MakeDaemonSet("test-pod", "").
				PodTemplateSpecManagedByKueue().
				Queue("new-test-queue").
				PodTemplateSpecQueue("new-test-queue").
				PodTemplateAnnotation(podconstants.SuspendedByParentAnnotation, FrameworkName).
				Obj(),
		},
		"daemonset without queue with pod template spec queue": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").PodTemplateSpecQueue("test-queue").Obj(),
			want:      testingdaemonset.MakeDaemonSet("test-pod", "").PodTemplateSpecQueue("test-queue").Obj(),
		},
		"LocalQueueDefaulting enabled, default lq is created, job doesn't have queue label": {
			localQueueDefaulting: true,
			defaultLqExist:       true,
			daemonset:            testingdaemonset.MakeDaemonSet("test-pod", "default").Obj(),
			want: testingdaemonset.MakeDaemonSet("test-pod", "default").
				PodTemplateSpecManagedByKueue().
				Queue("default").
				PodTemplateSpecQueue("default").
				PodTemplateAnnotation(podconstants.SuspendedByParentAnnotation, FrameworkName).
				Obj(),
		},
		"LocalQueueDefaulting enabled, default lq is created, job has queue label": {
			localQueueDefaulting: true,
			defaultLqExist:       true,
			daemonset:            testingdaemonset.MakeDaemonSet("test-pod", "").Queue("test-queue").Obj(),
			want: testingdaemonset.MakeDaemonSet("test-pod", "").
				PodTemplateSpecManagedByKueue().
				Queue("test-queue").
				PodTemplateSpecQueue("test-queue").
				PodTemplateAnnotation(podconstants.SuspendedByParentAnnotation, FrameworkName).
				Obj(),
		},
		"LocalQueueDefaulting enabled, default lq isn't created, job doesn't have queue label": {
			localQueueDefaulting: true,
			defaultLqExist:       false,
			daemonset:            testingdaemonset.MakeDaemonSet("test-pod", "").Obj(),
			want: testingdaemonset.MakeDaemonSet("test-pod", "").
				Obj(),
		},
		"daemonset with queue and priority class": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Label(constants.WorkloadPriorityClassLabel, "test").
				Obj(),
			want: testingdaemonset.MakeDaemonSet("test-pod", "").
				PodTemplateSpecManagedByKueue().
				Queue("test-queue").
				Label(constants.WorkloadPriorityClassLabel, "test").
				PodTemplateSpecQueue("test-queue").
				PodTemplateAnnotation(podconstants.SuspendedByParentAnnotation, FrameworkName).
				PodTemplateSpecLabel(constants.WorkloadPriorityClassLabel, "test").
				Obj(),
		},
		"daemonset with queue, priority class and pod template spec queue, priority class": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("new-test-queue").
				Label(constants.WorkloadPriorityClassLabel, "new-test").
				PodTemplateSpecQueue("test-queue").
				PodTemplateSpecLabel(constants.WorkloadPriorityClassLabel, "test").
				Obj(),
			want: testingdaemonset.MakeDaemonSet("test-pod", "").
				PodTemplateSpecManagedByKueue().
				Queue("new-test-queue").
				Label(constants.WorkloadPriorityClassLabel, "new-test").
				PodTemplateSpecQueue("new-test-queue").
				PodTemplateAnnotation(podconstants.SuspendedByParentAnnotation, FrameworkName).
				PodTemplateSpecLabel(constants.WorkloadPriorityClassLabel, "new-test").
				Obj(),
		},
		"daemonset without queue with pod template spec queue and priority class": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").
				PodTemplateSpecQueue("test-queue").
				PodTemplateSpecLabel(constants.WorkloadPriorityClassLabel, "test").
				Obj(),
			want: testingdaemonset.MakeDaemonSet("test-pod", "").
				PodTemplateSpecQueue("test-queue").
				PodTemplateSpecLabel(constants.WorkloadPriorityClassLabel, "test").
				Obj(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx, _ := utiltesting.ContextWithLog(t)
			features.SetFeatureGateDuringTest(t, features.LocalQueueDefaulting, tc.localQueueDefaulting)
			t.Cleanup(jobframework.EnableIntegrationsForTest(t, "pod"))
			builder := utiltesting.NewClientBuilder().
				WithObjects(
					utiltesting.MakeNamespace("default"),
					utiltesting.MakeNamespace(""),
				)
			client := builder.Build()
			cqCache := cache.New(client)
			queueManager := queue.NewManager(client, cqCache)
			if tc.defaultLqExist {
				if err := queueManager.AddLocalQueue(ctx, utiltesting.MakeLocalQueue("default", "default").
					ClusterQueue("cluster-queue").
					Obj()); err != nil {
					t.Fatalf("failed to create default local queue: %s", err)
				}
			}
			w := &Webhook{
				client: client,
				queues: queueManager,
			}

			if err := w.Default(ctx, tc.daemonset); err != nil {
				t.Errorf("failed to set defaults for v1/daemonset: %s", err)
			}
			if diff := cmp.Diff(tc.want, tc.daemonset); len(diff) != 0 {
				t.Errorf("Default() mismatch (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateCreate(t *testing.T) {
	testCases := map[string]struct {
		daemonset *appsv1.DaemonSet
		wantErr   error
		wantWarns admission.Warnings
	}{
		"without queue": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").Obj(),
		},
		"valid queue name": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Obj(),
		},
		"invalid queue name": {
			daemonset: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test/queue").
				Obj(),
			wantErr: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "metadata.labels[kueue.x-k8s.io/queue-name]",
				},
			}.ToAggregate(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Cleanup(jobframework.EnableIntegrationsForTest(t, "pod"))

			builder := utiltesting.NewClientBuilder()
			client := builder.Build()

			w := &Webhook{client: client}

			ctx, _ := utiltesting.ContextWithLog(t)

			warns, err := w.ValidateCreate(ctx, tc.daemonset)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.IgnoreFields(field.Error{}, "BadValue", "Detail")); diff != "" {
				t.Errorf("Unexpected error (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(warns, tc.wantWarns); diff != "" {
				t.Errorf("Expected different list of warnings (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	testCases := map[string]struct {
		oldDaemonSet *appsv1.DaemonSet
		newDaemonSet *appsv1.DaemonSet
		wantErr      error
		wantWarns    admission.Warnings
	}{
		"without queue (no changes)": {
			oldDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").Obj(),
			newDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").Obj(),
		},
		"without queue": {
			oldDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Obj(),
			newDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").Obj(),
			wantErr: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "metadata.labels[kueue.x-k8s.io/queue-name]",
				},
			}.ToAggregate(),
		},
		"with queue (no changes)": {
			oldDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Obj(),
			newDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Obj(),
		},
		"with queue": {
			oldDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").Obj(),
			newDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Obj(),
		},
		"with queue (invalid)": {
			oldDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test/queue").
				Obj(),
			newDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test/queue").
				Obj(),
			wantErr: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "metadata.labels[kueue.x-k8s.io/queue-name]",
				},
			}.ToAggregate(),
		},
		"with queue (number ready)": {
			oldDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				NumberReady(1).
				Obj(),
			newDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue-new").
				NumberReady(1).
				Obj(),
			wantErr: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "metadata.labels[kueue.x-k8s.io/queue-name]",
				},
			}.ToAggregate(),
		},
		"update priority-class": {
			oldDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Label(constants.WorkloadPriorityClassLabel, "test-1").
				Obj(),
			newDaemonSet: testingdaemonset.MakeDaemonSet("test-pod", "").
				Queue("test-queue").
				Label(constants.WorkloadPriorityClassLabel, "test-2").
				Obj(),
			wantErr: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "metadata.labels[kueue.x-k8s.io/priority-class]",
				},
			}.ToAggregate(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Cleanup(jobframework.EnableIntegrationsForTest(t, "pod"))

			builder := utiltesting.NewClientBuilder()
			client := builder.Build()

			w := &Webhook{client: client}

			ctx, _ := utiltesting.ContextWithLog(t)

			warns, err := w.ValidateUpdate(ctx, tc.oldDaemonSet, tc.newDaemonSet)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.IgnoreFields(field.Error{}, "BadValue", "Detail")); diff != "" {
				t.Errorf("Unexpected error (-want,+got):\n%s", diff)
			}
			if diff := cmp.Diff(warns, tc.wantWarns); diff != "" {
				t.Errorf("Expected different list of warnings (-want,+got):\n%s", diff)
			}
		})
	}
}