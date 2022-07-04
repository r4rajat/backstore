package controllers

import (
	"context"
	"fmt"
	"github.com/r4rajat/backstore/pkg/apis/backstore.github.com/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"os"
	"time"
)

type restoreController struct {
	client   dynamic.Interface
	informer cache.SharedIndexInformer
	queue    workqueue.RateLimitingInterface
	kClient  kubernetes.Interface
}

func NewRestoreController(client dynamic.Interface, dynInformer dynamicinformer.DynamicSharedInformerFactory, kClient kubernetes.Interface) *restoreController {
	queue := os.Getenv("BACKSTORE_RESTORE_QUEUE")
	if queue == "" {
		queue = "BACKSTORE"
	}
	inf := dynInformer.ForResource(schema.GroupVersionResource{
		Group:    "backstore.github.com",
		Version:  "v1alpha1",
		Resource: "restores",
	}).Informer()

	newBackupController := &restoreController{
		client:   client,
		informer: inf,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), queue),
		kClient:  kClient,
	}

	inf.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: newBackupController.addHandler,
		},
	)
	return newBackupController
}

func (rstr *restoreController) Run(ch <-chan struct{}) {
	fmt.Println("Starting Backup Controller")
	if !cache.WaitForCacheSync(ch, rstr.informer.HasSynced) {
		fmt.Print("waiting for cache to be synced\n")
	}
	go wait.Until(rstr.worker, 1*time.Second, ch)

	<-ch
}

func (rstr *restoreController) worker() {
	for rstr.processItem() {

	}
}

func (rstr *restoreController) processItem() bool {
	item, shutdown := rstr.queue.Get()
	if shutdown {
		return false
	}
	defer rstr.queue.Forget(item)
	defer rstr.queue.ShutDown()
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("Error getting key from cache.\nreason --> %s", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("Error splitting key into namespace and name.\nReason --> %s", err.Error())
		return false
	}
	fmt.Printf("Restore Name and Namespace are %s and %s", name, ns)
	err = rstr.createRestore(ns, name)
	if err != nil {
		return false
	}
	defer rstr.queue.Done(item)

	return true
}

func (rstr *restoreController) addHandler(obj interface{}) {
	fmt.Println("Restore Created")
	rstr.queue.Add(obj)
}

func (rstr *restoreController) createRestore(ns string, name string) error {
	restoreResource, err := rstr.client.Resource(schema.GroupVersionResource{
		Group:    "backstore.github.com",
		Version:  "v1alpha1",
		Resource: "restores",
	}).Namespace(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error Getting Backup Resource %s from Namespace %s.\nReason --> %s", name, ns, err.Error())
	}
	restore := v1alpha1.Restore{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(restoreResource.Object, &restore)
	if err != nil {
		log.Printf("Error Converting Unstructured Object to Structured Object.\nReason --> %s", err.Error())
	}
	log.Printf("Structured Object --> %v", restore)
	volumeSnapshotClass := restore.Spec.VolumeSnapshotClassName
	apiGroup := "snapshot.storage.k8s.io"
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: restore.Name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &volumeSnapshotClass,
			DataSource: &corev1.TypedLocalObjectReference{
				Name:     restore.Spec.BackupName,
				Kind:     "VolumeSnapshot",
				APIGroup: &apiGroup,
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(restore.Spec.Storage),
				},
			},
		},
	}

	_, err = rstr.kClient.CoreV1().PersistentVolumeClaims(restore.Namespace).Create(context.Background(), &pvc, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil

}
