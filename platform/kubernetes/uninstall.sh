#!/bin/bash

# Run this script to uninstall the platform from a kubernetes cluster

# reomve micro secrets
kubectl delete secret micro-secrets

# uninstall the resources
cd ./resource/cockroachdb;
bash uninstall.sh;
cd ../etcd;
bash uninstall.sh;
cd ../nats;
bash uninstall.sh;

# delete the PVs and PVCs
kubectl delete pvc --all
kubectl delete pv --all

# move to the /kubernetes folder and apply the deployments
cd ../..;
kubectl delete -f service

# go back to the top level
cd ..;
