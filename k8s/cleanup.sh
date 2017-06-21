kubectl delete -f k8s/statefulset.yml

kubectl get pvc | awk 'NR>1 {print "pvc/"$1}' | xargs kubectl delete

kubectl get pv | awk 'NR>1 {print "pv/"$1}' | xargs kubectl delete
