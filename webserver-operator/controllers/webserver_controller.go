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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	//"sigs.k8s.io/controller-runtime/pkg/log"

	mydomainv1 "github.com/zenkzen/kube-operators/webserver-operators/api/v1"
)

// WebserverReconciler reconciles a Webserver object
type WebserverReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=my.domain,resources=webservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.domain,resources=webservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=my.domain,resources=webservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Webserver object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *WebserverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := r.Log.WithValues("Webserver", req.NamespacedName)

	instance := &mydomainv1.Webserver{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			log.Log.Info("Webserver resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		log.Log.Error(err, "Failed to get Webserver")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	// Check if the webserver deployment already exists, if not, create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForWebserver(instance)
		log.Log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Log.Error(err, "Failed to get Deployment")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	// Ensure the deployment replicas and image are the same as the spec
	var replicas int32 = int32(instance.Spec.Replicas)
	image := instance.Spec.Image

	var needUpd bool
	if *found.Spec.Replicas != replicas {
		log.Log.Info("Deployment spec.replicas change", "from", *found.Spec.Replicas, "to", replicas)
		found.Spec.Replicas = &replicas
		needUpd = true
	}

	if (*found).Spec.Template.Spec.Containers[0].Image != image {
		log.Log.Info("Deployment spec.template.spec.container[0].image change", "from", (*found).Spec.Template.Spec.Containers[0].Image, "to", image)
		found.Spec.Template.Spec.Containers[0].Image = image
		needUpd = true
	}

	if needUpd {
		err = r.Update(ctx, found)
		if err != nil {
			log.Log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Check if the webserver service already exists, if not, create a new one
	foundService := &corev1.Service{
		/*
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "apps/v1",
			},
		*/
	}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name + "-service", Namespace: instance.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		srv := r.serviceForWebserver(instance)
		log.Log.Info("Creating a new Service", "Service.Namespace", srv.Namespace, "Service.Name", srv.Name)
		err = r.Create(ctx, srv)
		if err != nil {
			log.Log.Error(err, "Failed to create new Servie", "Service.Namespace", srv.Namespace, "Service.Name", srv.Name)
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Log.Error(err, "Failed to get Service")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	// Tbd: Ensure the service state is the same as the spec, your homework

	// reconcile webserver operator in again 10 seconds
	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebserverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mydomainv1.Webserver{}).
		Complete(r)
}

func (r *WebserverReconciler) deploymentForWebserver(ws *mydomainv1.Webserver) *appsv1.Deployment {
	labels := labelsForWebserver(ws.Name)
	var replicas int32 = int32(ws.Spec.Replicas)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ws.Name,
			Namespace: ws.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: ws.Spec.Image,
						Name:  "webserver",
						//Command: []string{"sleep", "360000"},
					}},
				},
			},
		},
	}
	// Set Webserver instance as the owner and controller
	//return dep, ctrl.SetControllerReference(ws, dep, r.Scheme)
	if err := ctrl.SetControllerReference(ws, dep, r.Scheme); err != nil {
		log.Log.Error(err, "Failed to SetControllerReference")
	}
	return dep
}

// serviceForWebserver returns a webserver-service service object
func (r *WebserverReconciler) serviceForWebserver(ws *mydomainv1.Webserver) *corev1.Service {
	labels := labelsForWebserver(ws.Name)
	srv := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ws.Name + "-service",
			Namespace: ws.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Protocol: corev1.ProtocolTCP,
					NodePort: 30010,
					Port:     80,
				},
			},
			Selector: map[string]string{
				"app":          "webserver",
				"webserver_cr": ws.Name,
			},
		},
	}

	// Set Webserver instance as the owner and controller
	//return dep, ctrl.SetControllerReference(ws, dep, r.Scheme)
	if err := ctrl.SetControllerReference(ws, srv, r.Scheme); err != nil {
		log.Log.Error(err, "Failed to SetControllerReference")
	}
	return srv
}

func labelsForWebserver(name string) map[string]string {
	return map[string]string{"app": "webserver", "webserver_cr": name}
}
