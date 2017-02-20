# Autohelm

Filesystem watcher and auto minikube deployer for Mac OS

## Requirements

minikube >= 0.14

docker configured to dockerd in minikube

kubectl configured to local minikube cluster


## Install

```curl -O ```

## Usage

Simple usage:

From your project root:
```autohelm -name chart-name -namespace services -configfile values.dev.yaml```

Supported flags:

```
kubeconfig - path to kube config (default ~/.kube/config)

name - helm chart name
helmdir - name of directory with main helm chart (it should be in root of your project)
configfile - helm configuration file (default values.yaml)
namespace - kubernetes namespace to deploy
tiller-namespace - kubernetes namespace with tiller (default kube-system)
d - delete chart after autohelm exit (In progress)
t - seconds to wait for rebuild (Default 5)
image-tag-name - name of variable with new image tag in helm chart

attach - attach to container after deploy or not (Works only if helm wait option works)
deploy - deployment name to attach with
container - container name to attach with
```

## Features

Key features:

- Watch your project files
- Automatic docker image rebuild
- Automatic helm deploy to your local minikube
- Automatic attach to particular container in new deployment
