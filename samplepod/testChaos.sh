#!/bin/bash
kubectl delete -f deploy/crds/chaos_v1alpha1_chaospod_cr.yaml
kubectl apply -f samplepod/someotherpod.yml
kubectl apply -f samplepod/somepod.yml
kubectl apply -f deploy/crds/chaos_v1alpha1_chaospod_cr.yaml 
