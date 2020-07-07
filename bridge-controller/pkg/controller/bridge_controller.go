package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/zmalik/kudo-bridge/bridge-controller/pkg/apis/kudobridge/v1alpha1"
	"github.com/zmalik/kudo-bridge/bridge-controller/pkg/client"
	"github.com/zmalik/kudo-bridge/bridge-controller/pkg/kudobridge/bridge"

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

	bridge *bridge.Bridge
}

func NewController(client *client.Client) *Controller {
	bridge := &bridge.Bridge{
		Client: client,
	}
	return &Controller{
		client:     client,
		bridge:     bridge,
		maxRetries: 1,
	}
}

func (c *Controller) Run(ctx context.Context) {
	stopCh := make(chan struct{})
	defer close(stopCh)
	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	c.informer = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return c.client.Bridge.KudobridgeV1alpha1().BridgeInstances("").List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return c.client.Bridge.KudobridgeV1alpha1().BridgeInstances("").Watch(context.TODO(), options)
			},
		},
		&v1alpha1.BridgeInstance{},
		0, //No resync
		cache.Indexers{},
	)

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldObj, _ := old.(*v1alpha1.BridgeInstance)
			newObj, _ := new.(*v1alpha1.BridgeInstance)
			if oldObj.GetResourceVersion() != newObj.GetResourceVersion() {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(newObj)
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
	obj, _, err := c.informer.GetStore().GetByKey(key)
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

	return c.bridge.Process(ro)
}
