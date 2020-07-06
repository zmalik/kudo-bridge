package watcher

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zmalik/kudo-bridge/crd-controller/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	uruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	client     *client.Client
	queue      workqueue.RateLimitingInterface
	informer   cache.SharedIndexInformer
	maxRetries int

	GroupVersion string
	Kind         string
	Namespace    string
}

func NewController(client *client.Client, groupVersion, kind, namespace string) *Controller {
	return &Controller{
		client:       client,
		GroupVersion: groupVersion,
		Kind:         kind,
		Namespace:    namespace,
	}
}

func (c *Controller) Run(ctx context.Context) {
	group, version, err := getGroupVersion(c.GroupVersion)
	if err != nil {
		uruntime.HandleError(fmt.Errorf("error parsing group version"))
		return
	}
	gvk := &schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    c.Kind,
	}
	meta, err := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(c.client.Discovery)).RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		log.Errorf("Cannot watch the provided CRD :%v", err)
		os.Exit(1)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	c.informer = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return c.client.Dynamic.Resource(meta.Resource).Namespace(c.Namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return c.client.Dynamic.Resource(meta.Resource).Namespace(c.Namespace).Watch(options)
			},
		},
		&unstructured.Unstructured{},
		0, //No resync
		cache.Indexers{},
	)

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldObj, _ := old.(*unstructured.Unstructured)
			newObj, _ := new.(*unstructured.Unstructured)
			if oldObj.GetResourceVersion() != newObj.GetResourceVersion() {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err == nil {
					c.queue.Add(key)
				}

			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
	})

	go c.informer.Run(stopCh)

	log.Infoln("Controller started.")
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		uruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	log.Infoln("Controller synced.")

	wait.Until(c.runWorker, time.Second, stopCh)
}
func getGroupVersion(groupVersion string) (string, string, error) {
	gv := strings.Split(groupVersion, "/")
	if len(gv) != 2 {
		return "", "", fmt.Errorf("error finding group and version in %s", groupVersion)
	}
	return gv[0], gv[1], nil
}

func (c *Controller) runWorker() {
	for c.processNext() {
	}
}

func (c *Controller) processNext() bool {
	key, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	if err == nil {
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < c.maxRetries {
		log.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		log.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		uruntime.HandleError(err)
	}
	return true
}

func (c *Controller) processItem(key string) error {
	obj, _, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}
	if obj == nil {
		return nil
	}
	ro, ok := obj.(runtime.Object)
	if !ok {
		return fmt.Errorf("object with key %s is not a runtime.Object", key)
	}

	return Process(c.client, ro)
}
