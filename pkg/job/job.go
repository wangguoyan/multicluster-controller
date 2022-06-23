package job

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github/wangguoyan/multicluster-controller/pkg/cluster"
	"github/wangguoyan/multicluster-controller/pkg/controller"
	"github/wangguoyan/multicluster-controller/pkg/manager"
	"k8s.io/client-go/rest"
	"sync"
	"time"
)

type WatchJob struct {
	cfg                *rest.Config
	resources          []*WatchResource
	logger             logr.Logger
	restClient         *rest.RESTClient
	clusterCfgCache    sync.Map
	clusterInfoCache   sync.Map
	clusterCancelCache sync.Map
	stopChan           chan struct{}
	synClusterOnce     sync.Once // 同步集群信息的动作只做一次
	hooks              hookList
}

func (w *WatchJob) execHooks(hooks []hook, info ClusterInfo, errs ...error) {
	if hooks == nil {
		return
	}
	var err error
	if errs != nil {
		err = errs[0]
	}
	for i := range hooks {
		hooks[i].handle(info, err)
	}
}
func (w *WatchJob) AddHookHandler(handlers ...HookHandler) *WatchJob {
	for i := range handlers {
		handler := handlers[i]
		if handler.Succeed != nil {
			w.hooks.succeeds = append(w.hooks.succeeds, hook{func(info ClusterInfo, err error) {
				handler.Succeed(info)
			}})
		}
		if handler.Failed != nil {
			w.hooks.failed = append(w.hooks.failed, hook{func(info ClusterInfo, err error) {
				handler.Failed(info, err)
			}})
		}
		if handler.Stop != nil {
			w.hooks.stop = append(w.hooks.stop, hook{func(info ClusterInfo, err error) {
				handler.Stop(info)
			}})
		}
	}

	return w
}

func NewWatchJob(cfg *rest.Config, res []*WatchResource, logger logr.Logger) (*WatchJob, error) {
	if len(res) == 0 {
		return nil, errors.New("watch resource is empty")
	}
	restClient, err := CreateClusterRESTClient(cfg)
	if err != nil {
		return nil, err
	}
	watchJob := &WatchJob{
		cfg:                cfg,
		resources:          res,
		logger:             logger,
		restClient:         restClient,
		clusterCfgCache:    sync.Map{},
		clusterCancelCache: sync.Map{},
		stopChan:           make(chan struct{}),
	}
	return watchJob, nil
}

// StartResourceWatch 启动集群监听，如果已经存在对应监听则忽略
func (w *WatchJob) StartResourceWatch(clusters ...ClusterInfo) error {
	for i := range clusters {
		cluInfo := clusters[i]
		_, ok := w.clusterCancelCache.Load(cluInfo.GetKey())
		if ok {
			continue
		}

		cancel, err := w.doResourceWatch(cluInfo)
		if err != nil {
			return err
		}
		w.clusterCancelCache.Store(cluInfo.GetKey(), cancel)
	}

	return nil
}

func (w *WatchJob) StopResourceWatch(clusters ...ClusterInfo) {
	for i := range clusters {
		cluInfo := clusters[i]
		cancel, ok := w.clusterCancelCache.Load(cluInfo.GetKey())
		if !ok {
			continue
		}
		(cancel.(context.CancelFunc))()
		w.clusterCancelCache.Delete(cluInfo.GetKey())
		w.execHooks(w.hooks.stop, cluInfo)
	}
}

func (w *WatchJob) StopAll() {
	w.clusterCancelCache.Range(func(key, value interface{}) bool {
		(value.(context.CancelFunc))()
		cluInfo, ok := w.clusterInfoCache.Load(key)
		if ok {
			w.execHooks(w.hooks.stop, cluInfo.(ClusterInfo))
		}
		return true
	})
	close(w.stopChan)
}

// 如果存在现有集群监听，则重启
func (w *WatchJob) restartResourceWatch(clusterInfo ClusterInfo) error {
	w.StopResourceWatch(clusterInfo)
	// 先启动新的监听
	newCancel, err := w.doResourceWatch(clusterInfo)
	if err != nil {
		return err
	}
	w.clusterCancelCache.Store(clusterInfo.GetKey(), newCancel)

	return err
}

// 创建并启动指定集群监听
func (w *WatchJob) doResourceWatch(clusterInfo ClusterInfo) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	// 因为mgr start方法为阻塞方法，所以需要启动协程执行
	go func() {
		clusterMgr := manager.New()
		cfg, err := w.getClusterCfg(clusterInfo)
		if err != nil {
			errChan <- err
			return
		}

		// 遍历需要监听的列表
		for i := range w.resources {
			resource := w.resources[i]

			co := controller.New(resource.Reconciler, controller.Options{})
			cl := cluster.New(clusterInfo.GetKey(), cfg, cluster.Options{})
			// client并注入到reconciler中
			delegatingClient, err := cl.GetDelegatingClient()
			if err != nil {
				errChan <- err
				return
			}
			resource.Reconciler.InjectClusterClient(*delegatingClient)
			if err = co.WatchResourceReconcileObject(ctx, cl, resource.ObjectType, resource.WatchOptions); err != nil {
				errChan <- err
				return
			}
			clusterMgr.AddController(co)
		}
		err = clusterMgr.Start(ctx)
		errChan <- err
	}()

	var err error
	// 等待一定时间后如果没有发生错误则认为执行成功
	select {
	case err = <-errChan:
		w.execHooks(w.hooks.failed, clusterInfo, err)
	case <-time.After(3 * time.Second):
	}
	if err == nil {
		w.execHooks(w.hooks.succeeds, clusterInfo)
	}
	return cancel, err
}

func (w *WatchJob) getClusterCfg(clusterInfo ClusterInfo) (*rest.Config, error) {
	config, ok := w.clusterCfgCache.Load(clusterInfo.GetKey())

	// 如果没有缓存，从集群直接获取
	if !ok {
		cfg, err := GetCfgByClusterInfo(clusterInfo)
		if err != nil {
			return nil, err
		}
		w.clusterCfgCache.Store(clusterInfo.GetKey(), cfg)
		return cfg, nil
	}
	return config.(*rest.Config), nil
}
