/*
Copyright 2022.

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

package controllers

import (
	"context"
	"errors"
	"fmt"
	myappv1 "github.com/freelizhun/hellocrd/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelloReconciler reconciles a Hello object
type HelloReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=myapp.freelizhun.com,resources=hellos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=myapp.freelizhun.com,resources=hellos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=myapp.freelizhun.com,resources=hellos/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Hello object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *HelloReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//logrus.Infof("Received request Namespace: %s and Name: %s", req.Namespace, req.Name)
	logger := ctrl.Log.WithValues("hello", req.NamespacedName)
	logger.Info(
		"Received request",
		"Namespace",
		req.Namespace,
		"Name",
		req.Name,
	)
	var hello myappv1.Hello
	if err := r.Get(ctx, req.NamespacedName, &hello); err != nil {
		logger.Error(err, "unable to get hello")
		return ctrl.Result{}, err
	}

	if hello.Status.Phase == "" {
		hello.Status.Phase = myappv1.HelloPending
	}
	//logrus.Infof("Check phase Phase: %s", hello.Status.Phase)
	logger.Info("Check phase", "Phase", hello.Status.Phase)
	requeue := false

	switch hello.Status.Phase {
	case myappv1.HelloPending:
		// Create a pod to run commands.
		pod := getHelloPod(&hello)
		if err := ctrl.SetControllerReference(&hello, pod, r.Scheme); err != nil {
			logger.Error(err, "Fail to set controller reference")
			return  ctrl.Result{}, err
		}
		if err := r.Create(ctx, pod); err != nil {
			logger.Error(err, "Fail to create pod")
			return ctrl.Result{}, err
		}
		hello.Status.Phase = myappv1.HelloRunning
	case myappv1.HelloRunning:
		pod := &corev1.Pod{}
		if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
			logger.Error(err, "Fail to get pod")
			return ctrl.Result{}, err
		}

		if pod.Status.Phase == corev1.PodSucceeded {
			hello.Status.Phase = myappv1.HelloSucceeded
		} else if pod.Status.Phase == corev1.PodFailed {
			hello.Status.Phase = myappv1.HelloFailed
		} else {
			requeue = true
		}
	case myappv1.HelloSucceeded:
		logger.Info("Have done!")
		return ctrl.Result{}, nil
	case myappv1.HelloFailed:
		pod := getHelloPod(&hello)
		if err := r.Delete(ctx, pod); err != nil {
			logger.Error(err, "Fail to delete pod")
			return ctrl.Result{}, err
		}
		hello.Status.Phase = myappv1.HelloPending
	default:
		logger.Error(
			nil,
			"Invalid phase",
			"Phase",
			hello.Status.Phase,
		)
		return ctrl.Result{}, errors.New("Invalid phase")
	}
	// Update hello status.
	if err := r.Status().Update(ctx, &hello); err != nil {
		logger.Error(err, "Fail to update hello status")
		return ctrl.Result{}, err
	}


	return ctrl.Result{Requeue: requeue}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myappv1.Hello{}).
		Complete(r)
}

func getHelloPod(hello *myappv1.Hello) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hello.Namespace,
			Name:      hello.Name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "ubuntu",
					Image: "busybox",
					Command: []string{
						"/bin/sh",
						"-c",
						fmt.Sprintf(
							"seq %d | xargs -I{} echo \"Hello\"",
							hello.Spec.HelloTimes,
						),
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}