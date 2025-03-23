/*
Copyright 2025 Rafal Jan.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ingressv1alpha1 "github.com/rafal-jan/ingress-duplicator/api/v1alpha1"
)

// AppIngressReconciler reconciles a AppIngress object
type AppIngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ingress.example.com,resources=appingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ingress.example.com,resources=appingresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ingress.example.com,resources=appingresses/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Condition Types for AppIngress
const (
	ConditionTypeNamespaceValid = "NamespaceValid"
	ConditionTypeIngressCreated = "IngressCreated"
)

// Finalizer for AppIngress cleanup
const (
	finalizerName = "ingress.example.com/cleanup"
)

// Reconcile handles the reconciliation loop for AppIngress resources
func (r *AppIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling AppIngress")

	// Get AppIngress
	appIngress := &ingressv1alpha1.AppIngress{}
	if err := r.Get(ctx, req.NamespacedName, appIngress); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !appIngress.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(appIngress, finalizerName) {
			logger.Info("Cleaning up associated Ingress", "namespace", appIngress.Spec.TargetNamespace)

			// Try to delete the Ingress
			ingress := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appIngress.Spec.Template.Name,
					Namespace: appIngress.Spec.TargetNamespace,
				},
			}
			if err := r.Delete(ctx, ingress); err != nil {
				if !apierrors.IsNotFound(err) {
					logger.Error(err, "Failed to delete Ingress during cleanup")
					return ctrl.Result{}, err
				}
				// If the Ingress is already gone, we can proceed with removing the finalizer
				logger.Info("Ingress already deleted or not found")
			}

			// Remove finalizer to allow AppIngress deletion
			controllerutil.RemoveFinalizer(appIngress, finalizerName)
			if err := r.Update(ctx, appIngress); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Cleanup completed successfully")
			return ctrl.Result{}, nil
		}
		// Finalizer already removed, nothing to do
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(appIngress, finalizerName) {
		controllerutil.AddFinalizer(appIngress, finalizerName)
		if err := r.Update(ctx, appIngress); err != nil {
			return ctrl.Result{}, err
		}
		// After adding finalizer, continue with reconciliation to set initial conditions
	}

	// Check if target namespace exists
	targetNs := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: appIngress.Spec.TargetNamespace}, targetNs); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "Target namespace not found", "namespace", appIngress.Spec.TargetNamespace)
			meta.SetStatusCondition(&appIngress.Status.Conditions, metav1.Condition{
				Type:    ConditionTypeNamespaceValid,
				Status:  metav1.ConditionFalse,
				Reason:  "NotFound",
				Message: "Target namespace does not exist",
			})
			if err := r.Status().Update(ctx, appIngress); err != nil {
				logger.Error(err, "Failed to update AppIngress status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Set namespace valid condition
	meta.SetStatusCondition(&appIngress.Status.Conditions, metav1.Condition{
		Type:    ConditionTypeNamespaceValid,
		Status:  metav1.ConditionTrue,
		Reason:  "Valid",
		Message: "Target namespace exists",
	})

	// Create/Update Ingress
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appIngress.Spec.Template.Name,
			Namespace: appIngress.Spec.TargetNamespace,
		},
	}

	// Create or update ingress - skip owner reference for cross-namespace objects
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		// Update ingress spec and metadata
		ingress.Labels = appIngress.Spec.Template.Labels
		ingress.Annotations = appIngress.Spec.Template.Annotations
		ingress.Spec = appIngress.Spec.Template.Spec
		return nil
	}); err != nil {
		logger.Error(err, "Failed to create/update Ingress")
		meta.SetStatusCondition(&appIngress.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeIngressCreated,
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: "Failed to create/update Ingress: " + err.Error(),
		})
		if err := r.Status().Update(ctx, appIngress); err != nil {
			logger.Error(err, "Failed to update AppIngress status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// Update success condition
	meta.SetStatusCondition(&appIngress.Status.Conditions, metav1.Condition{
		Type:    ConditionTypeIngressCreated,
		Status:  metav1.ConditionTrue,
		Reason:  "Created",
		Message: "Ingress created/updated successfully",
	})

	if err := r.Status().Update(ctx, appIngress); err != nil {
		logger.Error(err, "Failed to update AppIngress status")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciliation completed successfully")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ingressv1alpha1.AppIngress{}).
		Named("appingress").
		Complete(r)
}
