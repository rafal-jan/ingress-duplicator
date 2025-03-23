# Product Context

## Problem Statement
When using Kubernetes Ingress controllers, some implementations do not support referencing Services across namespaces. This limitation can be problematic in multi-team environments where:
- Services need to be kept in team-specific namespaces
- Ingress resources need to be managed centrally
- Platform teams need to control Ingress creation

## Solution
The AppIngress controller provides a solution by:
1. Allowing teams to keep their Services in their namespaces
2. Enabling platform teams to manage Ingress resources in designated namespaces
3. Automating the creation of Ingress resources in target namespaces
4. Ensuring proper validation of target namespaces

## User Experience Goals
- Simple, intuitive API similar to familiar Kubernetes resources
- Clear validation feedback
- Transparent status reporting
- Easy troubleshooting through standard kubectl commands
- Minimal configuration required

## Use Cases
1. **Centralized Ingress Management**
   - Platform team manages AppIngress CRs in a central namespace
   - Actual Ingress resources created in service namespaces

2. **Multi-Team Environments**
   - Teams work in isolated namespaces
   - Ingress configuration managed through templates
   - Cross-namespace service exposure

3. **Controlled Access**
   - Platform teams control Ingress creation
   - Teams maintain ownership of Services
   - Clear separation of responsibilities
