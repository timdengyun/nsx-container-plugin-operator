# NSX Container Plugin Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Overview

An operator for leveraging NSX as the default container networking solution for an
Kubernetes/Openshift cluster. The operator will be deployed in the early phases of
Openshift cluster deployment or after the kubectl is ready in Kubernetes cluster,
and it will take care of deploying NSX integration components, and precisely:

* The NSX container plugin (NCP) deployment
* The nsx-ncp-bootstrap daemonset
* The nsx-node-agent daemonset

The nsx-container-plugin operator monitors a dedicated ConfigMap, applies changes
to NCP and nsx-node-agent configuration, and creates/restarts the relevant pods
so that the relevant configuration changes are picked up.

The nsx-container-plugin operator also monitors the nsx-node-agent status and
updates the network status on relevant nodes.

In addition, the nsx-container-plugin operator is able to monitor nodes ensuring
the corresponding NSX logical port is enabled as a container host logical port.

For Openshift 4 clusters, the nsx-container-plugin operator especially monitors
the `network.config.openshift.io` CR to update the container network CIDRs used by NCP.

## Try it out

### Pull or Build Image

Pull the packed image for docker:
```
docker pull vmware/nsx-container-plugin-operator:latest
```

For containerd:
```
ctr image pull docker.io/vmware/nsx-container-plugin-operator:latest
```

Building the nsx-container-plugin operator is very simple. From the project root
directory simply type the following command, which based on docker build tool.

```
make all
```

At the moment the nsx-container-plugin operator only works on native Kubernetes
or Openshift 4 environments

### Deploy

#### Kubernetes

Edit the operator yaml files in `deploy/kubernetes` then apply them.

#### Openshift

Generate install-config.yaml by using openshift-install command.
```
$ openshift-install --dir=MY_CLUSTER create install-config
```

Edit `MY_CLUSTER/install-config.yaml` to update networking section.
Change `networkType` to `ncp`(case insensitive).
Set container network CIDRs `clusterNetwork` in `MY_CLUSTER/install-config.yaml`.

Create manifest files:
```
$ openshift-install --dir=MY_CLUSTER create manifests
```
Put operator yaml files from `deploy/openshift4/` to `MY_CLUSTER/manifests`,
edit configmap.yaml about operator configurations, add the operator image and
NCP image in operator.yaml.

Generate ignition configuration files:
```
$ openshift-install --dir=MY_CLUSTER create ignition-configs
```
This bootstrap ignition file will be added to the terraform tfvars.
Then use terraform to install Openshift 4 cluster on vSphere.

### Upgrade

For upgrading, all yaml files in `deploy/${platform}/` should be involved,
especially to check the `image` and `NCP_IMAGE` in `deploy/${platform}/operator.yaml


## Documentation

### Cluster network config (Openshift specific)
Cluster network config is initially set in install-config.yaml, user could apply
`network.config.openshift.io` CRD to update `clusterNetwork` in `manifests/cluster-network-02-config.yml`.
*Example configurations*
```
apiVersion: config.openshift.io/v1
kind: Network
metadata:
  name: cluster
spec:
  clusterNetwork:
  - cidr: 10.10.0.0/14
  networkType: ncp
```

### Operator ConfigMap

Operator ConfigMap `nsx-ncp-operator-config` is used to provide NCP configurations.
As for now we only support NSX Policy API, single Tier topology on Openshift 4,
single or two Tiers topology on native Kubernetes.

#### Kubernetes

Some fields are mandatory including `adaptor`, `cluster`, `nsx_api_managers`,
`container_ip_blocks`, `tier0_gateway`(for single T1 case), `top_tier_router`
(for single T0 case), `external_ip_pools`(for SNAT mode).. If any of above
options is not provided in the operator ConfigMap, the operator will fail to
reconcile configurations, error message swill be added in ncpinstall nsx-ncp
Degraded conditions

#### OpenShift

The operator sets `policy_nsxapi` as True, `single_tier_topology` as True.
In the ConfigMap, some fields are mandatory including `adaptor`, `cluster`, `nsx_api_managers`,
`tier0_gateway`(for single T1 case), `top_tier_router`(for single T0 case),
`external_ip_pools`(for SNAT mode). If any of above options is not provided in the
operator ConfigMap, the operator will fail to reconcile configurations, error messages
will be added in clusteroperator nsx-ncp Degraded conditions.

### NCP Image
User needs to set NCP image as an environment parameter `NCP_IMAGE` in `deploy/${platform}/operator.yaml`.

### Unsafe changes
* (Openshift specific) If CIDRs in `clusterNetwork` are already applied, it is
unsafe to remove them. NSX NCP operator won't fail when it detects some existing
network CIDRs are deleted, but the removal may cause unexpected issues.
* NSX NCP operator uses tags to mark the container host logical ports, deleting these tags
from NSX manager will cause network realization failure on corresponding nodes.

## Contributing

We welcome community contributions to the NSX Container plugin Operator!

Before you start working with nsx-container-plugin-operator, you should sign our
contributor license agreement (CLA).

If you wish to contribute code and you have not signed our CLA, our bot will update
the issue when you open a Pull Request.
For more detailed information, refer to [CONTRIBUTING.md](CONTRIBUTING.md).

For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq).

## License

This repository is available under the [Apache 2.0 license](LICENSE).

Add
