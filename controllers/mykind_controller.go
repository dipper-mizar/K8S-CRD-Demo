package controllers

import (
	mygroupv1 "K8S-CRD-Demo/api/v1"
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NewDeploy create a new deploy object
func NewDeploy(owner *mygroupv1.Mykind, logger logr.Logger, scheme *runtime.Scheme) map[string]*appsv1.Deployment {
	originOwnerName := owner.Name
	labelsMySQL := map[string]string{"app": originOwnerName + "-mysql"}
	selectorMySQL := &metav1.LabelSelector{MatchLabels: labelsMySQL}
	deployMySQL := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      originOwnerName + "-mysql",
			Namespace: owner.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: owner.Spec.ReplicasMySQL,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsMySQL,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            originOwnerName + "-mysql",
							Image:           owner.Spec.ImageMySQL,
							Ports:           []corev1.ContainerPort{{ContainerPort: owner.Spec.PortMySQL}},
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env:             owner.Spec.EnvsMySQL,
						},
					},
				},
			},
			Selector: selectorMySQL,
		},
	}
	// add ControllerReference for deployment
	if err := controllerutil.SetControllerReference(owner, deployMySQL, scheme); err != nil {
		msg := fmt.Sprintf("***SetControllerReference for Deployment %s/%s failed!***",
			owner.Namespace, originOwnerName+"-mysql")
		logger.Error(err, msg)
	}

	labelsCov := map[string]string{"app": originOwnerName + "-cov"}
	selectorCov := &metav1.LabelSelector{MatchLabels: labelsCov}
	deployCov := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      originOwnerName + "-cov",
			Namespace: owner.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: owner.Spec.ReplicasCov,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsCov,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            originOwnerName + "-cov",
							Image:           owner.Spec.ImageCov,
							Ports:           []corev1.ContainerPort{{ContainerPort: owner.Spec.PortCov}},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
				},
			},
			Selector: selectorCov,
		},
	}
	// add ControllerReference for deployment
	if err := controllerutil.SetControllerReference(owner, deployCov, scheme); err != nil {
		msg := fmt.Sprintf("***SetControllerReference for Deployment %s/%s failed!***",
			owner.Namespace, originOwnerName+"-cov")
		logger.Error(err, msg)
	}
	deployMap := map[string]*appsv1.Deployment{}
	deployMap["MySQL"] = deployMySQL
	deployMap["Cov"] = deployCov
	return deployMap
}

// NewService create a new service object
func NewService(owner *mygroupv1.Mykind, logger logr.Logger, scheme *runtime.Scheme) map[string]*corev1.Service {
	originOwnerName := owner.Name

	srvCov := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      originOwnerName + "-cov",
			Namespace: owner.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: owner.Spec.PortCov, NodePort: owner.Spec.NodeportCov}},
			Selector: map[string]string{
				"app": originOwnerName + "-cov",
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	// add ControllerReference for service
	if err := controllerutil.SetControllerReference(owner, srvCov, scheme); err != nil {
		msg := fmt.Sprintf("***setcontrollerReference for Service %s/%s failed!***", owner.Namespace,
			originOwnerName+"-cov")
		logger.Error(err, msg)
	}
	serviceMap := map[string]*corev1.Service{}
	serviceMap["Cov"] = srvCov
	return serviceMap
}

// MykindReconciler reconciles a Mykind object
type MykindReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mygroup.ips.com.cn,resources=mykinds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mygroup.ips.com.cn,resources=mykinds/status,verbs=get;update;patch
func (r *MykindReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	fmt.Println("---start Reconcile---")
	ctx := context.Background()
	lgr := r.Log.WithValues("mykind", req.NamespacedName)

	// your logic here
	/*1. create/update deploy
	  ========================*/
	mycrd_instance := &mygroupv1.Mykind{}
	if err := r.Get(ctx, req.NamespacedName, mycrd_instance); err != nil {
		lgr.Error(err, "***Get crd instance failed(maybe be deleted)! please check!***")
		return reconcile.Result{}, err
	}
	/*if mycrd_instance.DeletionTimestamp != nil {
	        lgr.Info("---Deleting crd instance,cleanup subresources---")
	        return reconcile.Result{}, nil
	}*/
	oldDeploy := &appsv1.Deployment{}
	newDeploy := NewDeploy(mycrd_instance, lgr, r.Scheme)
	for _, deploy := range newDeploy {
		if err := r.Get(ctx, req.NamespacedName, oldDeploy); err != nil && errors.IsNotFound(err) {
			lgr.Info("---Creating deploy---")
			// 1. create Deploy
			if err := r.Create(ctx, deploy); err != nil {
				lgr.Error(err, "***create deploy failed!***")
				return reconcile.Result{}, err
			}
			lgr.Info("---Create deploy done---")
		} else {
			if !reflect.DeepEqual(oldDeploy.Spec, deploy.Spec) {
				lgr.Info("---Updating deploy---")
				oldDeploy.Spec = deploy.Spec
				if err := r.Update(ctx, oldDeploy); err != nil {
					lgr.Error(err, "***Update old deploy failed!***")
					return reconcile.Result{}, err
				}
				lgr.Info("---Update deploy done---")
			}
		}
	}
	/*2. create/update Service
	  =========================*/
	oldService := &corev1.Service{}
	newService := NewService(mycrd_instance, lgr, r.Scheme)
	for _, service := range newService {
		if err := r.Get(ctx, req.NamespacedName, oldService); err != nil && errors.IsNotFound(err) {
			lgr.Info("---Creating service---")
			if err := r.Create(ctx, service); err != nil {
				lgr.Error(err, "***Create service failed!***")
				return reconcile.Result{}, err
			}
			lgr.Info("---Create service done---")
			return reconcile.Result{}, nil
		} else {
			if !reflect.DeepEqual(oldService.Spec, service.Spec) {
				lgr.Info("---Updating service---")
				clstip := oldService.Spec.ClusterIP //!!!clusterip unable be changed!!!
				oldService.Spec = service.Spec
				oldService.Spec.ClusterIP = clstip
				if err := r.Update(ctx, oldService); err != nil {
					lgr.Error(err, "***Update service failed!***")
					return reconcile.Result{}, err
				}
				lgr.Info("---Update service done---")
				return reconcile.Result{}, nil
			}
		}
	}
	lgr.Info("!!!err from Get maybe is nil,please check!!!")
	//end your logic
	return ctrl.Result{}, nil
}

func (r *MykindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1.Mykind{}).
		Complete(r)
}
