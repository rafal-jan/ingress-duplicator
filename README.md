# Ingress Duplicator

A Kubernetes controller that enables cross-namespace Ingress resource creation through a custom AppIngress resource. This solves the limitation of Ingress controllers that don't support cross-namespace Service references by allowing platform teams to manage Ingress resources centrally while creating them in target service namespaces.

## Features

- Create Ingress resources across namespaces using AppIngress custom resources
- Template-based Ingress specification similar to Deployment's Pod template pattern
- Automatic validation of target namespaces
- Status conditions for easy troubleshooting
- Automatic cleanup of created Ingress resources
- Clear separation between platform team management and service team namespaces

## Use Cases

1. **Centralized Ingress Management**
   - Platform teams manage AppIngress CRs in a central namespace
   - Actual Ingress resources are created in service namespaces

2. **Multi-Team Environments**
   - Teams work in isolated namespaces
   - Ingress configuration managed through templates
   - Cross-namespace service exposure

3. **Controlled Access**
   - Platform teams control Ingress creation
   - Teams maintain ownership of Services
   - Clear separation of responsibilities

## Installation

### Prerequisites
- Kubernetes v1.11.3+
- Go v1.24.1+
- Docker v17.03+
- kubectl v1.11.3+

### Deploy to Cluster

1. Build and push the controller image:
```sh
make docker-build docker-push IMG=<registry>/ingress-duplicator:tag
```

2. Install CRDs:
```sh
make install
```

3. Deploy the controller:
```sh
make deploy IMG=<registry>/ingress-duplicator:tag
```

> **NOTE**: You may need cluster-admin privileges to install CRDs and RBAC resources.

## Usage

1. Create an AppIngress resource:

```yaml
apiVersion: ingress.example.com/v1alpha1
kind: AppIngress
metadata:
  name: example-ingress
  namespace: platform-team
spec:
  targetNamespace: app-team
  template:
    metadata:
      name: app-ingress
      annotations:
        nginx.ingress.kubernetes.io/rewrite-target: /
    spec:
      rules:
      - host: app.example.com
        http:
          paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: app-service
                port:
                  number: 80
```

2. Verify the Ingress creation:
```sh
# Check AppIngress status
kubectl get appingress -n platform-team

# Verify Ingress creation in target namespace
kubectl get ingress -n app-team
```

### Status Conditions

The AppIngress resource reports status through conditions:

- `NamespaceValid`: Indicates if the target namespace exists
- `IngressCreated`: Shows the status of Ingress creation/updates

## Cleanup

1. Delete AppIngress resources:
```sh
kubectl delete -k config/samples/
```

2. Remove CRDs:
```sh
make uninstall
```

3. Uninstall the controller:
```sh
make undeploy
```

## Development

### Running Tests
```sh
# Run tests
make test

# Generate manifests
make manifests

# Generate code
make generate
```

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Known Limitations

- Cross-namespace owner references are not supported (Kubernetes limitation)
- Manual deletion required if finalizer is manually removed

## License

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
