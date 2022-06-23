/*
Copyright 2018 The Multicluster-Controller Authors.

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

package handler

import (
	"github/wangguoyan/multicluster-controller/pkg/reconcile"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type EnqueueRequestForObject struct {
	Context    string
	Queue      workqueue.Interface
	Filter     func(obj interface{}) bool
	Predicates []predicate.Predicate
}

func (e *EnqueueRequestForObject) enqueue(obj interface{}) {

	o, err := meta.Accessor(obj)
	if err != nil {
		return
	}

	r := reconcile.Request{Context: e.Context}
	r.Namespace = o.GetNamespace()
	r.Name = o.GetName()

	e.Queue.Add(r)
}

func (e *EnqueueRequestForObject) OnAdd(obj interface{}) {
	if !e.Filter(obj) {
		return
	}
	c := event.CreateEvent{}

	// Pull Object out of the object
	if o, ok := obj.(client.Object); ok {
		c.Object = o
	} else {
		return
	}
	for _, p := range e.Predicates {
		if !p.Create(c) {
			return
		}
	}
	e.enqueue(obj)
}

func (e *EnqueueRequestForObject) OnUpdate(oldObj, newObj interface{}) {
	if !e.Filter(newObj) {
		return
	}
	u := event.UpdateEvent{}

	if o, ok := oldObj.(client.Object); ok {
		u.ObjectOld = o
	} else {
		return
	}

	// Pull Object out of the object
	if o, ok := newObj.(client.Object); ok {
		u.ObjectNew = o
	} else {
		return
	}

	for _, p := range e.Predicates {
		if !p.Update(u) {
			return
		}
	}

	e.enqueue(newObj)
}

func (e *EnqueueRequestForObject) OnDelete(obj interface{}) {
	if !e.Filter(obj) {
		return
	}
	d := event.DeleteEvent{}

	var ok bool
	if _, ok = obj.(client.Object); !ok {
		// If the object doesn't have Metadata, assume it is a tombstone object of type DeletedFinalStateUnknown
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return
		}

		// Set obj to the tombstone obj
		obj = tombstone.Obj
	}

	// Pull Object out of the object
	if o, ok := obj.(client.Object); ok {
		d.Object = o
	} else {
		return
	}
	for _, p := range e.Predicates {
		if !p.Delete(d) {
			return
		}
	}
	e.enqueue(obj)
}
