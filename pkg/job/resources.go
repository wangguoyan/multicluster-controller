package job

import (
	"github/wangguoyan/multicluster-controller/pkg/controller"
	"github/wangguoyan/multicluster-controller/pkg/reconcile"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WatchResource 监听资源，包括类型和监听方法
type WatchResource struct {
	ObjectType   client.Object
	Scheme       *runtime.Scheme
	Reconciler   reconcile.Reconciler
	WatchOptions controller.WatchOptions
}

type ClusterInfo interface {
	GetApiServer() string
	GetToken() string
	GetKey() string
}
