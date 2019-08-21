# ARM Tag to Node Label Synchronization

Based off of a pre-existing document.

## Purpose

The purpose of this Kubernetes controller is to sync ARM VM/VMSS tags and node labels in an AKS cluster.
Users can choose whether to only sync ARM tags as node labels, sync node labels as ARM tags,
or perform a two-way sync.

## Motivation

Multiple customers have required this synchronization.
Their motivation is billing organization, housekeeping and overall resource tracking which works well on ARM tags.

## How it will work

### Kubernetes Configuration

- Default settings will have one way synchronization with VMSS tags as node labels.

- The controller runs as a deployment with 2 replicas. Leader election is enabled.

- The controller can be run with one of the following authentication methods:
    - Service Principals.
    - User Assigned Identity via "Pod Identity".

- Configurations can be specified in a Kubernetes ConfigMap. Configurable options include:
    - `syncDirection`: Direction of synchronization. Default is `arm-to-node`. Other options are `two-way` and `node-to-arm`.
    - `interval`: Configurable interval for synchronization.
    - `labelPrefix`: The node label prefix, with a default of `azure.tags`. An empty prefix will be permitted.
    - `resourceGroupFilter`: The controller can be limited to run on only nodes within a resource group filter (i.e. nodes that exist in RG1, RG2, RG3).
    - `conflictPolicy`: The policy for conflicting tag/label values. ARM tags or node labels can be given priority. ARM tags have priority by default (`arm-precedence`). Another option is to not update tags and raise Kubernetes event (`ignore`) and `node-precedence`. 

- Finished project will have sample YAML files for deployment, the options configmap, and managed identity will be provided with instructions on what to edit before applying to a cluster.

Sample configuration for options ConfigMap:

```
apiVersion: v1
kind: ConfigMap
metadata:
    name: tag-label-sync
    namespace: default
data:
    syncDirection: "arm-to-node"
    interval: "1"
    labelPrefix: "azure.tags"
    conflictPolicy: "arm-precedence"
    resourceGroupFilter: "none"
```

### Pseudo Code

For each VM/VMSS and node:
    - For any tag that exists on the VM/VMSS but does not exist as a label on the node, the label will be created, (and vice versa with labels and tags, if two-way sync is enabled).
    - If there is a conflict where a tag and label exist with the same name and a different value,
      the default action is that nothing will be done to resolve the conflict and the conflict will raise a Kubernetes
      event.
    - ARM tags will be added as node labels with configurable prefix, and a default prefix of `azure.tags`, with the form 
    `azure.tags/<tag-name>/<tag-value>`. This default prefix is to encourage the use of a prefix.
    - Node tags may not follow Azure tag name conventions (such as "kubernetes.io/os=linux" which contains '/'),
    so in that case...

## Implementation Challenges

- Currently, we need to wait for nodes to be ready to be able to run the controller and access VM/VMSS tags. This is not ideal.
- Cluster updates should not delete tags and labels.

## Possible Extensions

- Consider syncing node taints as ARM tags.

## Questions

- What is meant by a resource group filter? Won't the controller be run in a cluster with resources within a single resource group anyway?
- What kind of rules should be in place for conflicting tags/labels and strings that don't match naming rules when converted to a tag/label?
