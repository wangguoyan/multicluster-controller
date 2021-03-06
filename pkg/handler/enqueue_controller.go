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
	"github/wangguoyan/multicluster-controller/pkg/reference"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
)

type EnqueueRequestForController struct {
	Context           string
	ControllerContext string
	Queue             workqueue.Interface
	Predicate         func(obj interface{}) bool
}

func (e *EnqueueRequestForController) enqueue(obj interface{}) {
	if !e.Predicate(obj) {
		return
	}

	o, err := meta.Accessor(obj)
	if err != nil {
		return
	}

	// First, try to get a controller reference in the same cluster.
	if c := metav1.GetControllerOf(o); c != nil {
		r := reconcile.Request{Context: e.Context}
		r.Namespace = o.GetNamespace()
		r.Name = c.Name

		e.Queue.Add(r)
		return
	}

	// Then, try to get a multicluster controller reference.
	if c := reference.GetMulticlusterControllerOf(o); c != nil {
		if e.ControllerContext != "" && c.ClusterName != e.ControllerContext {
			return
		}
		r := reconcile.Request{Context: c.ClusterName}
		r.Namespace = c.Namespace
		r.Name = c.Name

		e.Queue.Add(r)
		return
	}
}

func (e *EnqueueRequestForController) OnAdd(obj interface{}) {
	e.enqueue(obj)
}

func (e *EnqueueRequestForController) OnUpdate(oldObj, newObj interface{}) {
	e.enqueue(newObj)
}

func (e *EnqueueRequestForController) OnDelete(obj interface{}) {
	e.enqueue(obj)
}
