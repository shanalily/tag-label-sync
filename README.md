# ARM Tags to Node Labels Controller

Create an AKS cluster and set current context to use this cluster.

Add the following environment variables to work with Azure authorization.
<!-- https://github.com/Azure-Samples/azure-sdk-for-go-samples -->


```
export AZURE_SUBSCRIPTION_ID=
export AZURE_TENANT_ID=
export AZURE_CLIENT_ID=
export AZURE_CLIENT_SECRET=

```

For MSI authentication: https://github.com/Azure/aad-pod-identity

Create ConfigMap with configurable options and apply to cluster.

Run `make` to build, then `make run` to run.


The deployment file is config/manager/manager.yaml. You can change sync-period to configure
min interval between reconciliation.
