/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT)
	defer cancel()

	var kubeconfig *string
	var masterUrl *string
	var namespace *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	masterUrl = flag.String("master_url", "https://localhost:8443", "master url to connect to cluster")
	namespace = flag.String("namespace", "", "namespace to list data from")

	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags(*masterUrl, *kubeconfig)
	if err != nil {
		log.Fatalf("error building config from master_url=%s and kubeconfig=%s", *masterUrl, *kubeconfig)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error building client from master_url=%s and kubeconfig=%s", *masterUrl, *kubeconfig)
	}
	pods, err := clientset.CoreV1().Pods(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("error listing pods: %s", err)
	}
	log.Printf("There are %d pods in the cluster", len(pods.Items))
	wr := tabwriter.NewWriter(os.Stdout, 10, 25, 0, ' ',
		tabwriter.StripEscape)
	fmt.Fprint(wr, "№\tName\tNode\t\n")
	for i := range pods.Items {
		_, err = fmt.Fprintf(wr, "%v\t%s\t%s\n",
			1+i, pods.Items[i].Name, pods.Items[i].Status.HostIP,
		)
		if err != nil {
			log.Fatalf("error writing data: %s", err)
		}
	}
	err = wr.Flush()
	if err != nil {
		log.Fatalf("error writing data: %s", err)
	}
	nodePorts, err := clientset.CoreV1().Services(*namespace).List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{Kind: "node_port"},
		Limit:    0,
	})
	if err != nil {
		log.Fatalf("error listing node ports: %s", err)
	}
	log.Printf("There are %d NodePorts in the cluster", len(nodePorts.Items))

	fmt.Fprint(wr, "№\tName\tProtocol\tPort\tTarget\tNodePort\tIPs\t\n")
	for i := range nodePorts.Items {
		for j := range nodePorts.Items[i].Spec.Ports {
			_, err = fmt.Fprintf(wr, "%v\t%s\t%s\t%v\t%v\t%v\t%s\t\n",
				1+i+j,
				nodePorts.Items[i].Name,
				nodePorts.Items[i].Spec.Ports[j].Protocol,
				nodePorts.Items[i].Spec.Ports[j].Port,
				nodePorts.Items[i].Spec.Ports[j].TargetPort.String(),
				nodePorts.Items[i].Spec.Ports[j].NodePort,
				strings.Join(nodePorts.Items[i].Spec.ExternalIPs, ","),
			)
		}
	}
	err = wr.Flush()
	if err != nil {
		log.Fatalf("error writing data: %s", err)
	}

	// Examples for error handling:
	// - Use helper functions like e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	//namespace := "default"
	//pod := "example-xxxxx"
	//_, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
	//if errors.IsNotFound(err) {
	//	fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	//} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
	//	fmt.Printf("Error getting pod %s in namespace %s: %v\n",
	//		pod, namespace, statusError.ErrStatus.Message)
	//} else if err != nil {
	//	panic(err.Error())
	//} else {
	//	fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	//}
}
