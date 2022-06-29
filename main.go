package main

import (
	"github.com/r4rajat/backstore/pkg/controllers"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"log"
	"time"
)

func main() {
	dynamicClient, exssClientSet, kClient, err := newClient()
	if err != nil {
		log.Fatalf("Error Creating Dynamic Client.\nReason --> %s", err.Error())
	}
	infFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 10*time.Minute)
	backupController := controllers.NewBackupController(dynamicClient, infFactory, exssClientSet)
	restoreController := controllers.NewRestoreController(dynamicClient, infFactory, kClient)
	ch := make(<-chan struct{})
	go infFactory.Start(ch)
	backupController.Run(ch)
	restoreController.Run(ch)
}
