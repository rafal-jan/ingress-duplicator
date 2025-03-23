# Technical Context

## Development Environment
- Go 1.24.1
- kubebuilder v4.5.1

## Project Structure
- Generated using kubebuilder scaffold
- API Group: ingress.example.com
- API Version: v1alpha1
- Custom Resource: AppIngress
  - Spec:
    - template: Full Ingress template with metadata and spec
      - metadata: ObjectMeta for target Ingress
      - spec: Full IngressSpec configuration
    - targetNamespace: Target namespace for Ingress creation
  - Status:
    - conditions: Standard Kubernetes conditions
      - NamespaceValid: Target namespace existence check
      - IngressCreated: Ingress creation/update status

## Testing Environment
### Manual Testing
- Local Kubernetes cluster with Docker Desktop
- Manual testing workflow:
  1. Install CRDs with `make install`
  2. Run controller locally with `make run`
  3. Apply sample AppIngress resources
  4. Verify cross-namespace Ingress creation

### Automated Testing
- EnvTest framework for controller tests
- Test environment configuration:
  - Auto-created control plane
  - In-memory API server
  - KUBEBUILDER_ASSETS for required binaries
- Test organization:
  - Ginkgo/Gomega BDD-style tests
  - Shared test resources with Ordered container
  - BeforeAll for test namespace setup

## Key Files
- `api/v1alpha1/appingress_types.go`: AppIngress CRD definition with IngressTemplate type
- `internal/controller/appingress_controller.go`: Controller implementation
  - Reconciliation logic for namespace validation
  - Ingress creation/update with owner references
  - Status conditions management (NamespaceValid, IngressCreated)
  - RBAC annotations for required permissions
- `config/crd/bases/ingress.example.com_appingresses.yaml`: Generated CRD manifest
- `config/samples/ingress_v1alpha1_appingress.yaml`: Sample CR

## Dependencies
- sigs.k8s.io/controller-runtime v0.20.2
  - Controller runtime framework
  - Client abstractions
  - Manager and reconciler patterns
- k8s.io/apimachinery v0.32.1
  - Core type definitions
  - Meta types for conditions
- k8s.io/client-go v0.32.1
  - Kubernetes API access
  - Error types
- k8s.io/api v0.32.1
  - Ingress and Namespace types

## Build System
- Makefile-based build system
- Generated with kubebuilder
- Key targets:
  - `make manifests`: Generate CRDs and RBAC
  - `make generate`: Generate code
  - `make test`: Run tests
