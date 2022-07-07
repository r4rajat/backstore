package main

import (
	"flag"
	exss "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func newClient() (dynamic.Interface, exss.Interface, kubernetes.Interface, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Printf("Unable to get Home Directory of Current User.\nReason --> %s", err.Error())
	}
	kubeconfigPath := homeDir + "/.kube/config"
	log.Printf("Setting default kubeconfig location to --> %s", kubeconfigPath)
	kubeconfig := flag.String("kubeconfig", kubeconfigPath, "Location to your kubeconfig file")
	if (*kubeconfig != "") && (*kubeconfig != kubeconfigPath) {
		log.Printf("Recieved new kubeconfig location --> %s", *kubeconfig)
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Not able to create kubeconfig object from default location.\nReason --> %s", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("Not able to create kubeconfig object from inside pod.\nReason --> %s", err.Error())
		} else {
			log.Printf("In Cluater Config Created")
		}
	}
	log.Println("Created config object with provided kubeconfig")
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error Creating Dynamic Client.\nReason --> %s", err.Error())
	}
	exssClientSet, err := exss.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error occurred while creating External Snapshot Client Set with provided config.\nReason --> %s", err.Error())
	}
	kClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error occurred while creating Client Set with provided config.\nReason --> %s", err.Error())
	}
	return dynamicClient, exssClientSet, kClient, err
}
