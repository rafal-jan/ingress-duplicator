# System Patterns

## Architecture
- Kubernetes Custom Resource Definition (CRD) based
- Controller pattern implementation using controller-runtime
- Watch-based reconciliation loop

## Design Patterns
1. Kubernetes Controller Pattern
   - Watch for AppIngress resource changes
   - Reconcile desired state (Ingress in target namespace)
   - Status updates to reflect current state

2. Template Pattern
   - AppIngress uses template-based specification
   - Similar to Deployment's PodTemplate pattern
   - Allows for flexible Ingress configuration

3. Cross-Namespace Resource Management
   - Target namespace validation
   - Cross-namespace resource creation
   - Ownership and garbage collection considerations

## Controller Workflow
1. Watch AppIngress resources
2. On change:
   - Validate target namespace exists
   - Create/update Ingress in target namespace
   - Update status to reflect current state
3. Handle deletions and updates

## Error Handling
- Namespace validation errors:
  - Sets NamespaceValid condition to False
  - Records error in status with NotFound reason
- Resource creation/update errors:
  - Sets IngressCreated condition to False
  - Records detailed error message in status
- Status update failures:
  - Logged with error details
  - Returns error for requeue

## Status Conditions
- NamespaceValid:
  - True: Target namespace exists and is valid
  - False: Target namespace does not exist
- IngressCreated:
  - True: Ingress successfully created/updated
  - False: Failed to create/update Ingress

## RBAC Patterns
- Controller service account with:
  - AppIngress CRUD and status update permissions
  - Ingress CRUD permissions in target namespaces
  - Namespace read permissions for validation
  - Finalizer update permissions

## Resource Ownership and Cross-Namespace Management
- Cross-namespace owner references are not supported by Kubernetes
- Ingresses are created in target namespace without owner references
- Automatic cleanup using finalizers when AppIngress is deleted
- Finalizer "ingress.example.com/cleanup" ensures proper garbage collection

## Known Limitations
- Cross-namespace owner references not supported (by design)
- Manual deletion only if finalizer is manually removed

## Testing Patterns
### Controller Tests
- Ordered test setup for shared resources
  * Target namespace created once in BeforeAll
  * Tests use shared namespace to avoid EnvTest limitations with namespace deletion

### Test Cases Coverage
1. Namespace Validation
   - Test ingress creation with existing namespace
   - Test error handling with non-existent namespace

2. Ingress Management
   - Verify ingress creation with template spec
   - Verify ingress updates (host, rules, etc.)
   - Verify status condition updates

3. Test Structure
   - Separate contexts for different scenarios
   - BeforeEach for test-specific resource setup
   - AfterEach for test cleanup (except namespace)
   - Status condition verification for all operations
