package controllers

import (
	"context"
	"fmt"
	snapshots "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	exss "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	"github.com/r4rajat/backstore/pkg/apis/backstore.github.com/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"os"
	"time"
)

type backupController struct {
	client     dynamic.Interface
	informer   cache.SharedIndexInformer
	queue      workqueue.RateLimitingInterface
	exssClient exss.Interface
}

func NewBackupController(client dynamic.Interface, dynInformer dynamicinformer.DynamicSharedInformerFactory, exssClient exss.Interface) *backupController {
	queue := os.Getenv("BACKSTORE_BACKUP_QUEUE")
	if queue == "" {
		queue = "BACKSTORE"
	}
	inf := dynInformer.ForResource(schema.GroupVersionResource{
		Group:    "backstore.github.com",
		Version:  "v1alpha1",
		Resource: "backups",
	}).Informer()

	newBackupController := &backupController{
		client:     client,
		informer:   inf,
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), queue),
		exssClient: exssClient,
	}

	inf.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: newBackupController.addHandler,
		},
	)
	return newBackupController
}

func (bkup *backupController) Run(ch <-chan struct{}) {
	fmt.Println("Starting Backup Controller")
	if !cache.WaitForCacheSync(ch, bkup.informer.HasSynced) {
		fmt.Print("waiting for cache to be synced\n")
	}

	go wait.Until(bkup.worker, 1*time.Second, ch)

	<-ch
}

func (bkup *backupController) worker() {
	for bkup.processItem() {

	}
}

func (bkup *backupController) processItem() bool {
	item, shutdown := bkup.queue.Get()
	if shutdown {
		return false
	}
	defer bkup.queue.Forget(item)
	defer bkup.queue.ShutDown()

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("Error getting key from cache.\nreason --> %s", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("Error splitting key into namespace and name.\nReason --> %s", err.Error())
		return false
	}
	fmt.Printf("Backup Name and Namespace are %s and %s", name, ns)
	err = bkup.createBackup(ns, name)
	if err != nil {
		return false
	}
	defer bkup.queue.Done(item)

	return true
}

func (bkup *backupController) addHandler(obj interface{}) {
	fmt.Println("Backup Created")
	bkup.queue.Add(obj)
}

func (bkup *backupController) createBackup(ns string, name string) error {
	backupResource, err := bkup.client.Resource(schema.GroupVersionResource{
		Group:    "backstore.github.com",
		Version:  "v1alpha1",
		Resource: "backups",
	}).Namespace(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error Getting Backup Resource %s from Namespace %s.\nReason --> %s", name, ns, err.Error())
	}
	backup := v1alpha1.Backup{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(backupResource.Object, &backup)
	if err != nil {
		log.Printf("Error Converting Unstructured Object to Structured Object.\nReason --> %s", err.Error())
	}
	log.Printf("Structured Object --> %v", backup)
	volumeSnapshotClass := backup.Spec.VolumeSnapshotClassName
	snapshot := snapshots.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backup.Spec.VolumeSnapshotName,
			Namespace: backup.Spec.Namespace,
		},
		Spec: snapshots.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &volumeSnapshotClass,
			Source: snapshots.VolumeSnapshotSource{
				PersistentVolumeClaimName: &backup.Spec.PVC,
			},
		},
	}
	_, err = bkup.exssClient.SnapshotV1().VolumeSnapshots(backup.Spec.Namespace).Create(context.Background(), &snapshot, metav1.CreateOptions{})
	log.Printf("Snapshot Created for %s", name)
	if err != nil {
		return err
	}
	return nil
}
