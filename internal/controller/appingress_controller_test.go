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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	ingressv1alpha1 "github.com/rafal-jan/ingress-duplicator/api/v1alpha1"
)

var _ = Describe("AppIngress Controller", Ordered, func() {
	const (
		resourceName = "test-appingress"
		namespace    = "default"
		targetNs     = "test-target-namespace"
	)

	var (
		ctx                  = context.Background()
		namespacedName       = types.NamespacedName{Name: resourceName, Namespace: namespace}
		appIngress           *ingressv1alpha1.AppIngress
		controllerReconciler *AppIngressReconciler
	)

	BeforeAll(func() {
		// Create target namespace that will be used by tests that need it
		targetNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: targetNs,
			},
		}
		Expect(k8sClient.Create(ctx, targetNamespace)).To(Succeed())
	})

	Context("When target namespace does not exist", func() {
		BeforeEach(func() {
			nonExistentNs := "non-existent-namespace"
			// Create AppIngress instance
			appIngress = &ingressv1alpha1.AppIngress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: ingressv1alpha1.AppIngressSpec{
					Template: ingressv1alpha1.IngressTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-ingress",
						},
					},
					TargetNamespace: nonExistentNs,
				},
			}
			Expect(k8sClient.Create(ctx, appIngress)).To(Succeed())

			// Create controller instance
			controllerReconciler = &AppIngressReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
		})

		AfterEach(func() {
			if appIngress != nil {
				// Delete the AppIngress
				_ = k8sClient.Delete(ctx, appIngress)
				// Reconcile to handle the finalizer
				_, err := controllerReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())
				appIngress = nil
			}
		})

		It("should set appropriate conditions", func() {
			// Trigger reconciliation
			result, err := controllerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify conditions
			updatedAppIngress := &ingressv1alpha1.AppIngress{}
			Expect(k8sClient.Get(ctx, namespacedName, updatedAppIngress)).To(Succeed())

			nsCondition := findCondition(updatedAppIngress.Status.Conditions, ConditionTypeNamespaceValid)
			Expect(nsCondition).NotTo(BeNil())
			Expect(nsCondition.Status).To(Equal(metav1.ConditionFalse))
			Expect(nsCondition.Reason).To(Equal("NotFound"))
		})
	})

	Context("When target namespace exists", func() {
		BeforeEach(func() {
			// Create AppIngress instance
			appIngress = &ingressv1alpha1.AppIngress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: ingressv1alpha1.AppIngressSpec{
					Template: ingressv1alpha1.IngressTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-ingress",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &[]networkingv1.PathType{networkingv1.PathTypePrefix}[0],
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "test-service",
															Port: networkingv1.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					TargetNamespace: targetNs,
				},
			}
			Expect(k8sClient.Create(ctx, appIngress)).To(Succeed())

			// Create controller instance
			controllerReconciler = &AppIngressReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
		})

		AfterEach(func() {
			if appIngress != nil {
				// Delete the AppIngress
				_ = k8sClient.Delete(ctx, appIngress)
				// Reconcile to handle the finalizer
				_, err := controllerReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())
				appIngress = nil
			}
		})

		It("should create ingress in target namespace", func() {
			// Trigger reconciliation
			result, err := controllerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify ingress creation
			createdIngress := &networkingv1.Ingress{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      appIngress.Spec.Template.Name,
				Namespace: targetNs,
			}, createdIngress)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdIngress.Spec).To(Equal(appIngress.Spec.Template.Spec))

			// Verify conditions
			updatedAppIngress := &ingressv1alpha1.AppIngress{}
			Expect(k8sClient.Get(ctx, namespacedName, updatedAppIngress)).To(Succeed())
			Expect(updatedAppIngress.Status.Conditions).To(HaveLen(2))

			nsCondition := findCondition(updatedAppIngress.Status.Conditions, ConditionTypeNamespaceValid)
			Expect(nsCondition).NotTo(BeNil())
			Expect(nsCondition.Status).To(Equal(metav1.ConditionTrue))

			ingressCondition := findCondition(updatedAppIngress.Status.Conditions, ConditionTypeIngressCreated)
			Expect(ingressCondition).NotTo(BeNil())
			Expect(ingressCondition.Status).To(Equal(metav1.ConditionTrue))
		})

		It("should update existing ingress", func() {
			// First reconciliation to create ingress
			_, err := controllerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Update AppIngress template
			updatedHost := "updated-example.com"
			updatedAppIngress := &ingressv1alpha1.AppIngress{}
			Expect(k8sClient.Get(ctx, namespacedName, updatedAppIngress)).To(Succeed())
			updatedAppIngress.Spec.Template.Spec.Rules[0].Host = updatedHost
			Expect(k8sClient.Update(ctx, updatedAppIngress)).To(Succeed())

			// Second reconciliation to update ingress
			_, err = controllerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify ingress update
			updatedIngress := &networkingv1.Ingress{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      appIngress.Spec.Template.Name,
				Namespace: targetNs,
			}, updatedIngress)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedIngress.Spec.Rules[0].Host).To(Equal(updatedHost))
		})
	})

	Context("When AppIngress is deleted", func() {
		BeforeEach(func() {
			// Create AppIngress instance
			appIngress = &ingressv1alpha1.AppIngress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: ingressv1alpha1.AppIngressSpec{
					Template: ingressv1alpha1.IngressTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-ingress",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "example.com",
								},
							},
						},
					},
					TargetNamespace: targetNs,
				},
			}
			Expect(k8sClient.Create(ctx, appIngress)).To(Succeed())

			// Create controller instance
			controllerReconciler = &AppIngressReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// First reconciliation to create ingress and add finalizer
			result, err := controllerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		AfterEach(func() {
			if appIngress != nil {
				// Delete the AppIngress
				_ = k8sClient.Delete(ctx, appIngress)
				// Reconcile to handle the finalizer
				_, err := controllerReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())
				appIngress = nil
			}
		})

		It("should add finalizer on creation", func() {
			createdAppIngress := &ingressv1alpha1.AppIngress{}
			Expect(k8sClient.Get(ctx, namespacedName, createdAppIngress)).To(Succeed())
			Expect(createdAppIngress.Finalizers).To(ContainElement("ingress.example.com/cleanup"))
		})

		It("should cleanup ingress and remove finalizer on deletion", func() {
			// Verify ingress exists
			createdIngress := &networkingv1.Ingress{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      appIngress.Spec.Template.Name,
				Namespace: targetNs,
			}, createdIngress)
			Expect(err).NotTo(HaveOccurred())

			// Delete AppIngress
			Expect(k8sClient.Delete(ctx, appIngress)).To(Succeed())

			// Trigger reconciliation
			result, err := controllerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify ingress is deleted
			deletedIngress := &networkingv1.Ingress{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      appIngress.Spec.Template.Name,
				Namespace: targetNs,
			}, deletedIngress)
			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsNotFound(err)).To(BeTrue())

			// Verify finalizer is removed
			deletedAppIngress := &ingressv1alpha1.AppIngress{}
			err = k8sClient.Get(ctx, namespacedName, deletedAppIngress)
			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})

		It("should handle deletion when ingress is already gone", func() {
			// Delete ingress manually first
			ingress := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appIngress.Spec.Template.Name,
					Namespace: targetNs,
				},
			}
			Expect(k8sClient.Delete(ctx, ingress)).To(Succeed())

			// Delete AppIngress
			Expect(k8sClient.Delete(ctx, appIngress)).To(Succeed())

			// Trigger reconciliation
			result, err := controllerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify AppIngress is fully deleted
			deletedAppIngress := &ingressv1alpha1.AppIngress{}
			err = k8sClient.Get(ctx, namespacedName, deletedAppIngress)
			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
	})
})

// Helper function to find a condition by type
func findCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
