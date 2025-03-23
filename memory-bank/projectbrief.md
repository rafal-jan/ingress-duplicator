# Project Brief

## Overview
Building a kubernetes controller that will provide a custom resource for creating an Ingress in a different namespace.

## Core Features
- Provide `AppIngress` custom resource
  - Ingress resource template in `.spec.template` similarly to Deployment cotaining Pod template
  - Create the Ingress resource from the template in target namespace defined in `.spec.targetNamespace`
  - Validate target namespace existence before creation

## Target Users
Platform engineers using Ingress Controllers without support for cross-namespace Service references

## Technical Preferences (optional)
- Go 1.24.1
- kubebuilder