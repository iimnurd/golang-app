## How to Install using helm

Helm install -f helm-dev-env.yaml golang_app ./golang


### Upgrade 
Helm upgrade -f helm-dev-env.yaml golang_app ./golang


FLow :
outside request -> nginx-ingress kubernetes -> golang app service -> golang app (pod)