KubePort
========================

Utility to extract pods and services information from Kubernets deployment.
It reveals connection string for any `NodePort` type service matching regex

```shell
$ kubeports -h
Usage of ./build/kubeports:
  -grep string
    	regex to filter pods and services
  -kubeconfig string
    	(optional) absolute path to the kubeconfig file (default "/home/vodolaz095/.kube/config")
  -master_url string
    	master url to connect to cluster (default "https://localhost:8443")
  -namespace string
    	namespace to list data from


$ kubeports --master_url=https://192.168.39.243:8443 --grep=nginx
2025/05/08 14:14:03 KubePorts v0.1.6.6c8813c-Linux-x86_64 ivory_on_2025-05-08_11:13:21. Please, report bugs here: https://github.com/vodolaz095/kubeports/issues
2025/05/08 14:14:03 Dialing https://192.168.39.243:8443...
2025/05/08 14:14:03 Nodes found:
Name          Type        Address         
minikube      InternalIP  192.168.39.243  
minikube      Hostname    minikube        
minikube-m02  InternalIP  192.168.39.101  
minikube-m02  Hostname    minikube-m02    

2025/05/08 14:14:03 Podes matching `nginx` are found:
Name                              NodeIP    
nginx-deployment-647677fc66-mmth6 192.168.39.243
nginx-deployment-647677fc66-v4vmx 192.168.39.243

2025/05/08 14:14:03 Services matching `nginx` are found:
Name          Protocol  Port      Target    NodePort  Type      
nginx-service TCP       80        80        31080     NodePort  

2025/05/08 14:14:03 Writing connection strings for exposed by NodePort services
Name           Protocol  Connection           
nginx-service  TCP       192.168.39.243:31080 
nginx-service  TCP       192.168.39.101:31080 

2025/05/08 14:14:03 Goodbye!

```
