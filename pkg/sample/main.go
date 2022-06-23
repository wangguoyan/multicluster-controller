package main

import (
	"fmt"
	"github.com/go-logr/logr"
	"github/wangguoyan/multicluster-controller/pkg/job"
	reconcile2 "github/wangguoyan/multicluster-controller/pkg/reconcile"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type test struct {
}

func (t test) GetApiServer() string {
	return ""
}

func (t test) GetToken() string {
	return ""
}

func (t test) GetKey() string {
	return ""
}

func (t test) name() {

}

func main() {
	watchResources := []*job.WatchResource{
		{
			ObjectType: &corev1.Pod{},
			Reconciler: &testReconciler{},
		},
	}
	watchJob, err := job.NewWatchJob(ctrl.GetConfigOrDie(), watchResources, logr.DiscardLogger{})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 监听指定集群
	err = watchJob.StartResourceWatch(&test{})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	time.Sleep(5 * time.Second)
	watchJob.StopAll()
}

type testReconciler struct {
	cli client.Client
}

func (r *testReconciler) Reconcile(request reconcile2.Request) (reconcile.Result, error) {
	//TODO implement me
	return reconcile.Result{}, nil
}

func (r *testReconciler) InjectClusterClient(cli client.Client) {
	r.cli = cli
}
