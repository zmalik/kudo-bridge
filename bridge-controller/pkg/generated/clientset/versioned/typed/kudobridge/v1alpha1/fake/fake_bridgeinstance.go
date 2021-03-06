/*

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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/zmalik/kudo-bridge/bridge-controller/pkg/apis/kudobridge/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBridgeInstances implements BridgeInstanceInterface
type FakeBridgeInstances struct {
	Fake *FakeKudobridgeV1alpha1
	ns   string
}

var bridgeinstancesResource = schema.GroupVersionResource{Group: "kudobridge.dev", Version: "v1alpha1", Resource: "bridgeinstances"}

var bridgeinstancesKind = schema.GroupVersionKind{Group: "kudobridge.dev", Version: "v1alpha1", Kind: "BridgeInstance"}

// Get takes name of the bridgeInstance, and returns the corresponding bridgeInstance object, and an error if there is any.
func (c *FakeBridgeInstances) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.BridgeInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(bridgeinstancesResource, c.ns, name), &v1alpha1.BridgeInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BridgeInstance), err
}

// List takes label and field selectors, and returns the list of BridgeInstances that match those selectors.
func (c *FakeBridgeInstances) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.BridgeInstanceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(bridgeinstancesResource, bridgeinstancesKind, c.ns, opts), &v1alpha1.BridgeInstanceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BridgeInstanceList{ListMeta: obj.(*v1alpha1.BridgeInstanceList).ListMeta}
	for _, item := range obj.(*v1alpha1.BridgeInstanceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested bridgeInstances.
func (c *FakeBridgeInstances) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(bridgeinstancesResource, c.ns, opts))

}

// Create takes the representation of a bridgeInstance and creates it.  Returns the server's representation of the bridgeInstance, and an error, if there is any.
func (c *FakeBridgeInstances) Create(ctx context.Context, bridgeInstance *v1alpha1.BridgeInstance, opts v1.CreateOptions) (result *v1alpha1.BridgeInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(bridgeinstancesResource, c.ns, bridgeInstance), &v1alpha1.BridgeInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BridgeInstance), err
}

// Update takes the representation of a bridgeInstance and updates it. Returns the server's representation of the bridgeInstance, and an error, if there is any.
func (c *FakeBridgeInstances) Update(ctx context.Context, bridgeInstance *v1alpha1.BridgeInstance, opts v1.UpdateOptions) (result *v1alpha1.BridgeInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(bridgeinstancesResource, c.ns, bridgeInstance), &v1alpha1.BridgeInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BridgeInstance), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeBridgeInstances) UpdateStatus(ctx context.Context, bridgeInstance *v1alpha1.BridgeInstance, opts v1.UpdateOptions) (*v1alpha1.BridgeInstance, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(bridgeinstancesResource, "status", c.ns, bridgeInstance), &v1alpha1.BridgeInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BridgeInstance), err
}

// Delete takes name of the bridgeInstance and deletes it. Returns an error if one occurs.
func (c *FakeBridgeInstances) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(bridgeinstancesResource, c.ns, name), &v1alpha1.BridgeInstance{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBridgeInstances) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(bridgeinstancesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.BridgeInstanceList{})
	return err
}

// Patch applies the patch and returns the patched bridgeInstance.
func (c *FakeBridgeInstances) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.BridgeInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(bridgeinstancesResource, c.ns, name, pt, data, subresources...), &v1alpha1.BridgeInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BridgeInstance), err
}
