# Ordered Chaos Monkey Operator
This k8 operator is created for demonstrating purpose using the [operator-sdk](https://github.com/operator-framework/operator-sdk) in `v0.11.0`. 
It introduces a new Custom Ressource (Definition) "ChaosPod". 
You can define a prefix (prefixtokill) for to be killed pods.
As soon as you create/apply a chaos pod, it kills all pods with said prefix.

### Build & Publish Operator (on dockerhub)
```
operator-sdk build dxjoke/ordered-chaos-monkey-operator:v0.0.1
sed -i 's|REPLACE_IMAGE|dxjoke/ordered-chaos-monkey-operator:v0.0.1|g' deploy/operator.yaml
docker push dxjoke/ordered-chaos-monkey-operator:v0.0.1
```

### Run locally (instead of Build & Publish)
`export OPERATOR_NAME=ordered-chaos-monkey-operator`

Without debugging:

`operator-sdk up local --namespace=default`

With debugging:

`operator-sdk up local --namespace=default --enable-delve`


### Create Chaospod Operator
```
kubectl create -f deploy/crds/chaos_v1alpha1_chaospod_crd.yaml
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/operator.yaml
```


### Test/Create chaos pod
First lets create a to killable pod using:

`kubectl apply -f samplepod/somepod.yml`

Create the chaos pod (which should kill the pod of `somepod.yml`) using:

`kubectl apply -f deploy/crds/chaos_v1alpha1_chaospod_cr.yaml`


Inspect chaospod:
`kubectl describe chaospod`
You will see in Status/Killedpodnames a list of killed pod names (in our case "tokilltwocontainerspod"). 

Pod tokilltwocontainerspod shouldnt be running anymore as well.
