# Operator Manifest Verification
Operator Manifest Verification is a library that provides functions to verify the operator manifest bundles. These bundles are an amalagamation of [Operator-Lifecycle-Manager's](https://github.com/operator-framework/operator-lifecycle-manager) (OLM) [ClusterServiceVersion](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/design/building-your-csv.md) type, [CustomResourceDefinitions](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/), and [Package Manifest](https://github.com/operator-framework/operator-lifecycle-manager#discovery-catalogs-and-automated-upgrades) yamls.

Currently, this library reports errors and/or warnings for missing mandatory and optional fields, respectively. It also supports validation of `ClusterServiceVersion` yaml for any mismatched data types with Operator-Lifecycle-Manager's `ClusterServiceVersion` [type](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/pkg/api/apis/operators/v1alpha1/clusterserviceversion_types.go#L359:6). 

# Getting Started
The Operator Manifest Verfication library defines a single definition of a valid operator. It helps in validating operator manifest bundles before deploying them on cluster, and thus, helping in the operator development process.

# Usage
Currently, you can use this library with a command line tool:

## Command Line Tool
You can install the `operator-verify` tool from source using:

`$ go install`

`$ echo $PATH`

If you do not have `$(go env GOPATH)/bin` in your `$PATH`, 

`$ export PATH=$PATH:$(go env GOPATH)/bin`

This adds your workspace's bin subdirectory to your PATH. As a result, you can use the `operator-verify` tool anywhere on your system. Otherwise, you would have to `cd` to your workspace's `bin` directory to run the executable. 

To verify your ClusterServiceVersion yaml,

`$ operator-verify verify /path/to/filename.yaml`