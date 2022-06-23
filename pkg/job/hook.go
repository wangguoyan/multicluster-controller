package job

type hookList struct {
	succeeds []hook
	failed   []hook
	stop     []hook
}

type hook struct {
	handle func(info ClusterInfo, err error)
}
type HookHandler struct {
	Succeed func(info ClusterInfo)
	Failed  func(info ClusterInfo, err error)
	Stop    func(info ClusterInfo)
}
