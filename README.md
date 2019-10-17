# Ordered Chaos Monkey Operator
This k8 operator is created for demonstrating purpose using the [operator-sdk](https://github.com/operator-framework/operator-sdk) in `v0.11.0`. 
It introduces a new Custom Ressource (Definition) "ChaosPod". 
You can define a prefix (prefixtokill) for to be killed pods.
As soon as you create/apply a chaos pod, it kills all pods with said prefix.


### Run locally (instead of Build & Publish)
`export OPERATOR_NAME=ordered-chaos-monkey-operator`

Without debugging:

`operator-sdk up local --namespace=default`

With debugging:

`operator-sdk up local --namespace=default --enable-delve`

### Create Custom Resource
After the operator is running you can create a chaospod, for example with:

`kubectl create -f deploy/crds/chaos.wessner.me_v1alpha1_chaospod_cr.yaml`


### Build & Publish Operator (on dockerhub)
```
operator-sdk build dxjoke/ordered-chaos-monkey-operator:v0.0.1
sed -i 's|REPLACE_IMAGE|dxjoke/ordered-chaos-monkey-operator:v0.0.1|g' deploy/operator.yaml
docker push dxjoke/ordered-chaos-monkey-operator:v0.0.1
```

## Create Ordered Chaos Monkey Operator
Register CRD:

`kubectl create -f deploy/crds/chaos.wessner.me_chaospods_crd.yaml`

Create RBAC and Operator:
```
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/operator.yaml
```

Afterwards you can follow [Create Custom Resource](#create-custom-resource)

### Test/Create chaos pod
First lets create a to killable pod using:

`kubectl apply -f samplepod/somepod.yml`

Create the chaos pod (which should kill the pod of `somepod.yml`) using:

`kubectl apply -f deploy/crds/chaos_v1alpha1_chaospod_cr.yaml`


Inspect chaospod:

`kubectl get chaospod example-chaospod -o yaml`

```yaml
...
status:
  killedPodNames:
    df28d187-817d-4f6a-98a7-fbccf58aefa9: tokilltwocontainerspod
```


You will see a list of killed pod names (in our case "tokilltwocontainerspod"). 

Pod `tokilltwocontainerspod` shouldnt be running anymore as well.



### Remove Ordered Chaos Monkey Operator

```
kubectl delete -f deploy/crds/chaos.wessner.me_v1alpha1_chaospod_cr.yaml
kubectl delete -f deploy/crds/chaos.wessner.me_chaospods_crd.yaml
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/service_account.yaml
```