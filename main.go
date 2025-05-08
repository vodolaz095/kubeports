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
	"regexp"
	"sort"
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

type exposed struct {
	Name      string
	Port      int32
	Protocol  string
	Addresses []string
}

func main() {
	var ret []exposed
	var nodeIPs []string
	var err error
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
	var grep *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	masterUrl = flag.String("master_url", "https://localhost:8443", "master url to connect to cluster")
	namespace = flag.String("namespace", "", "namespace to list data from")
	grep = flag.String("grep", "", "regex to filter pods and services")
	flag.Parse()

	var match *regexp.Regexp
	if *grep != "" {
		match, err = regexp.Compile(*grep)
		if err != nil {
			log.Fatalf("error compiling regex %s: %s", *grep, err)
		}
	} else {
		match, _ = regexp.Compile(`.`)
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags(*masterUrl, *kubeconfig)
	if err != nil {
		log.Fatalf("error building config from master_url=%s and kubeconfig=%s", *masterUrl, *kubeconfig)
	}
	log.Printf("Dialing %s...", config.Host)
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error building client from master_url=%s and kubeconfig=%s", *masterUrl, *kubeconfig)
	}
	wr := tabwriter.NewWriter(os.Stdout, 10, 35, 1, ' ',
		tabwriter.StripEscape)

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		Limit: 0,
	})
	if err != nil {
		log.Fatalf("error listing node ports: %s", err)
	}
	log.Printf("Nodes found:")
	fmt.Fprint(wr, "№\tName\tType\tAddress\t\n")
	for i := range nodes.Items {
		for j := range nodes.Items[i].Status.Addresses {
			_, err = fmt.Fprintf(wr, "%v \t%s \t%s \t%s \t\n",
				i+1,
				nodes.Items[i].Name,
				nodes.Items[i].Status.Addresses[j].Type,
				nodes.Items[i].Status.Addresses[j].Address,
			)
			if nodes.Items[i].Status.Addresses[j].Type == "InternalIP" {
				nodeIPs = append(nodeIPs, nodes.Items[i].Status.Addresses[j].Address)
			}
		}
	}
	err = wr.Flush()
	if err != nil {
		log.Fatalf("error writing data: %s", err)
	}

	// reveal pods
	pods, err := clientset.CoreV1().Pods(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("error listing pods: %s", err)
	}
	log.Printf("Podes matching `%s` are found:", match.String())
	fmt.Fprint(wr, "№\tName\tNodeIP\t\n")
	for i := range pods.Items {
		if !match.MatchString(pods.Items[i].Name) {
			continue
		}
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

	// reveal services
	services, err := clientset.CoreV1().Services(*namespace).List(ctx, metav1.ListOptions{
		Limit: 0,
	})
	if err != nil {
		log.Fatalf("error listing node ports: %s", err)
	}
	log.Printf("Services matching `%s` are found:", match.String())

	fmt.Fprint(wr, "№\tName\tProtocol\tPort\tTarget\tNodePort\tType\t\n")
	for i := range services.Items {
		if !match.MatchString(services.Items[i].Name) {
			continue
		}
		for j := range services.Items[i].Spec.Ports {
			_, err = fmt.Fprintf(wr, "%v\t%s \t%s\t%v\t%v\t%v\t%s\t\n",
				1+i+j,
				services.Items[i].Name,
				services.Items[i].Spec.Ports[j].Protocol,
				services.Items[i].Spec.Ports[j].Port,
				services.Items[i].Spec.Ports[j].TargetPort.String(),
				services.Items[i].Spec.Ports[j].NodePort,
				services.Items[i].Spec.Type,
			)
			if err != nil {
				log.Fatalf("error writing to buffer: %s", err)
			}
			if services.Items[i].Spec.Ports[j].NodePort != 0 && services.Items[i].Spec.Type == "NodePort" {
				ret = append(ret, exposed{
					Name:      services.Items[i].Name,
					Protocol:  string(services.Items[i].Spec.Ports[j].Protocol),
					Port:      services.Items[i].Spec.Ports[j].NodePort,
					Addresses: nodeIPs,
				})
			}
		}
	}
	err = wr.Flush()
	if err != nil {
		log.Fatalf("error writing data: %s", err)
	}

	log.Printf("Writing connection strings for exposed by NodePort services")
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name > ret[j].Name
	})
	fmt.Fprint(wr, "№\tName\tProtocol\tConnection\t\n")
	for i := range ret {
		for j := range ret[i].Addresses {
			_, err = fmt.Fprintf(wr, "%v \t%s \t%s \t%s:%v\t\n",
				1+i+j,
				ret[i].Name,
				ret[i].Protocol,
				ret[i].Addresses[j],
				ret[i].Port,
			)
			if err != nil {
				log.Fatalf("error writing to buffer: %s", err)
			}
		}
	}
	err = wr.Flush()
	if err != nil {
		log.Fatalf("error writing data: %s", err)
	}
}
