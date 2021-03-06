/*
Copyright 2018 The Kubernetes Authors.

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

package client

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

// client is a client.Client that reads and writes directly from/to an API server.  It lazily initializes
// new clients at the time they are used, and caches the client.
type typedClient struct {
	cache      clientCache
	paramCodec runtime.ParameterCodec
}

// Create implements client.Client
func (c *typedClient) Create(ctx context.Context, obj runtime.Object, opts ...CreateOptionFunc) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	createOpts := &CreateOptions{}
	createOpts.ApplyOptions(opts)
	return o.Post().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Body(obj).
		VersionedParams(createOpts.AsCreateOptions(), c.paramCodec).
		Context(ctx).
		Do().
		Into(obj)
}

// Update implements client.Client
func (c *typedClient) Update(ctx context.Context, obj runtime.Object, opts ...UpdateOptionFunc) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	updateOpts := &UpdateOptions{}
	updateOpts.ApplyOptions(opts)
	return o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		Body(obj).
		VersionedParams(updateOpts.AsUpdateOptions(), c.paramCodec).
		Context(ctx).
		Do().
		Into(obj)
}

// Delete implements client.Client
func (c *typedClient) Delete(ctx context.Context, obj runtime.Object, opts ...DeleteOptionFunc) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	deleteOpts := DeleteOptions{}
	return o.Delete().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		Body(deleteOpts.ApplyOptions(opts).AsDeleteOptions()).
		Context(ctx).
		Do().
		Error()
}

// Patch implements client.Client
func (c *typedClient) Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOptionFunc) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}

	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	patchOpts := &PatchOptions{}
	return o.Patch(patch.Type()).
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), c.paramCodec).
		Body(data).
		Context(ctx).
		Do().
		Into(obj)
}

// Get implements client.Client
func (c *typedClient) Get(ctx context.Context, key ObjectKey, obj runtime.Object) error {
	r, err := c.cache.getResource(obj)
	if err != nil {
		return err
	}
	return r.Get().
		NamespaceIfScoped(key.Namespace, r.isNamespaced()).
		Resource(r.resource()).
		Context(ctx).
		Name(key.Name).Do().Into(obj)
}

// List implements client.Client
func (c *typedClient) List(ctx context.Context, obj runtime.Object, opts ...ListOptionFunc) error {
	r, err := c.cache.getResource(obj)
	if err != nil {
		return err
	}
	listOpts := ListOptions{}
	listOpts.ApplyOptions(opts)
	return r.Get().
		NamespaceIfScoped(listOpts.Namespace, r.isNamespaced()).
		Resource(r.resource()).
		VersionedParams(listOpts.AsListOptions(), c.paramCodec).
		Context(ctx).
		Do().
		Into(obj)
}

// UpdateStatus used by StatusWriter to write status.
func (c *typedClient) UpdateStatus(ctx context.Context, obj runtime.Object) error {
	o, err := c.cache.getObjMeta(obj)
	if err != nil {
		return err
	}
	// TODO(droot): examine the returned error and check if it error needs to be
	// wrapped to improve the UX ?
	// It will be nice to receive an error saying the object doesn't implement
	// status subresource and check CRD definition
	return o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.isNamespaced()).
		Resource(o.resource()).
		Name(o.GetName()).
		SubResource("status").
		Body(obj).
		Context(ctx).
		Do().
		Into(obj)
}
