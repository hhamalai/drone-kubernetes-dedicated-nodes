# Admission webhook for dedicated Drone CI runners

This projects setups admission webhook, that adds nodeSelector and toleration definitions
for Drone CI jobs. Together with tainted CI workers, this can be used to create host separation
with your CI jobs and rest of your workloads.


### Installation
Setup admission controller & webhook. Kubernetes manifest templates are
provided. There is also deploy.sh script that does full shell scripted setup for AWS using AWS Secrets Manager
and Kuberentes aws-secret-operator, which are used to store and synchronize the self signed
certificates and keys in AWS and into Kubernetes.

This script can be executed as follows
```
cd deploy
aws-vault exec "<replace-aws-profile>" -- sh build.sh registry.url/drone-pod-admission-controller:dev
```

