# ARM Tag to Node Label Synchronization

Based off of [this document](https://microsoft.sharepoint.com/:w:/r/teams/azurecontainercompute/_layouts/15/Doc.aspx?sourcedoc=%7B3a2d20bc-7fa4-450c-8bcf-67156b7b594d%7D&action=edit&wdPid=14896249).

## Purpose

The purpose of this Kubernetes controller is to sync ARM VM/VMSS tags and node labels in an AKS cluster.
Users can choose whether to only sync ARM tags as node labels, sync node labels as ARM tags,
or perform a two-way sync.

## Motivation

Multiple customers have required this synchronization.
Their motivation is billing organization, housekeeping and overall resource tracking which works well on ARM tags.

## How it will work

Default settings will have two-way synchronization with VMSS tags and node labels.

1. For each VMSS and node, any tags that exist on the VMSS and not as a label on the node, the label will be created,
and vice versa.
    - If there is a conflict where a tag and label exist with the same name and a different value,
      the default action is that nothing will be done to resolve the conflict and the conflict will be logged.
    - ARM tags will be added as node labels with configurable prefix, and a default prefix of `azure.tags`, with the form 
    `azure.tags/<tag-name>/<tag-value>`. This default prefix is to encourage the use of a prefix.
2. The controller runs as a deployment with 2 replicas. Leader election is enabled.
3. The controller can be run with one of the following authentication methods:
    - Service Principals.
    - User Assigned Identity via "Pod Identity".
4. The controller can be limited to run on only nodes within a resource group filter.
5. Configurable options include:
    - Switching to one-way synchronization.
    - Sychronizing VM tags instead of VMSS tags.
    - The node label prefix. An empty prefix will be permitted.
    - The policy for conflicting tags. VM/VMSS tags or node labels can be given priority.

## Implementation Challenges

- Currently, we need to wait for nodes to be ready to be able to run the controller and access VM/VMSS tags.

- Cluster updates should not delete tags and labels.

## Extensions

Consider syncing node taints as ARM tags.
