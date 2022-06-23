package job

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

func CreateClusterRESTClient(cfg *rest.Config) (*rest.RESTClient, error) {
	setConfigDefaults(cfg)
	return rest.RESTClientFor(cfg)
}

func setConfigDefaults(config *rest.Config) {
	gv := v1alpha1.GroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	//config.NegotiatedSerializer = v1alpha1.Codecs.WithoutConversion()
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
}

func GetCfgByClusterInfo(clusterInfo ClusterInfo) (*rest.Config, error) {
	tlsClientConfig := rest.TLSClientConfig{
		Insecure: true,
	}
	cfg := rest.Config{
		Host:            clusterInfo.GetApiServer(),
		TLSClientConfig: tlsClientConfig,
		BearerToken:     clusterInfo.GetToken(),
	}

	return &cfg, nil
}
