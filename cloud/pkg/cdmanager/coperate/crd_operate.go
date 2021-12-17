package coperate

import (
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	"io/ioutil"

	"sigs.k8s.io/yaml"

	v1alpha2api "github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdeviceapi/client/v1alpha2"
	v1alpha2 "github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdeviceapi/v1alpha2"

	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	restclient "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type deviceClient struct {
	crdClient *restclient.RESTClient
}

type OriginOptions struct {
	ConfigFile string
}

// KubeAPIConfig indicates the configuration for interacting with k8s server
type KubeAPIConfig struct {
	// Master indicates the address of the Kubernetes API server (overrides any value in KubeConfig)
	// such as https://127.0.0.1:8443
	// default ""
	// Note: Can not use "omitempty" option,  It will affect the output of the default configuration file
	Master string `json:"master"`
	// ContentType indicates the ContentType of message transmission when interacting with k8s
	// default "application/vnd.kubernetes.protobuf"
	ContentType string `json:"contentType,omitempty"`
	// QPS to while talking with kubernetes apiserve
	// default 100
	QPS int32 `json:"qps,omitempty"`
	// Burst to use while talking with kubernetes apiserver
	// default 200
	Burst int32 `json:"burst,omitempty"`
	// KubeConfig indicates the path to kubeConfig file with authorization and master location information.
	// default "/root/.kube/config"
	// +Required
	KubeConfig string `json:"kubeConfig"`
}

const (
	// Config
	DefaultKubeContentType = "application/vnd.kubernetes.protobuf"
	DefaultKubeConfig      = "/root/.kube/config"
	DefaultKubeQPS         = 100.0
	DefaultKubeBurst       = 200
)

// ConnectionConfig
type ConnectionConfig struct {
	KubeAPIConfig *KubeAPIConfig `json:"kubeAPIConfig,omitempty"`
}

func FromConfigFile(PATH string) *OriginOptions {
	return &OriginOptions{
		ConfigFile: PATH,
	}
}

func NewConnectionConfig() *ConnectionConfig {
	c := &ConnectionConfig{
		KubeAPIConfig: &KubeAPIConfig{
			Master:      "",
			ContentType: DefaultKubeContentType,
			QPS:         DefaultKubeQPS,
			Burst:       DefaultKubeBurst,
			KubeConfig:  DefaultKubeConfig,
		},
	}
	return c
}

func ParseConfig(filename string) (*ConnectionConfig, error) {
	cfg := NewConnectionConfig()

	cfg, err := Parse(filename, cfg)
	if err != nil {
		print("Parse error, err:", err)
	}

	return cfg, err
}

func Parse(filename string, cfg *ConnectionConfig) (*ConnectionConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		klog.Errorf("Failed to read configfile %s: %v", filename, err)
		return nil, err
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		klog.Errorf("Failed to unmarshal configfile %s: %v", filename, err)
		return nil, err
	}
	return cfg, nil
}

func InitRestClient(config *KubeAPIConfig) *v1alpha2api.DevicesV1alpha2Client {
	if config.KubeConfig == "" {
		print("Load /root/.kube/config failed.")
	}
	print("config.Master is :", config.Master)
	print("config Kubeconfig is :", config.KubeConfig)
	clientConfig, err := clientcmd.BuildConfigFromFlags(config.Master, config.KubeConfig)
	if err != nil {
		print("Error to bild config, err:", err)
	}

	clientConfig.QPS = float32(config.QPS)
	clientConfig.Burst = int(config.Burst)
	clientConfig.ContentType = runtime.ContentTypeProtobuf
	print("clientConfig user: ", clientConfig.Username, clientConfig.UserAgent)
	crdConfig := restclient.CopyConfig(clientConfig)
	crdConfig.ContentType = runtime.ContentTypeJSON
	print("client connection config user is :", crdConfig.Username)
	DeviceClient, err := NewForConfig(crdConfig)
	if err != nil {
		print("Error when transfer crdconfig to DeviceClient")
	}
	return DeviceClient
}

func NewForConfig(c *restclient.Config) (*v1alpha2api.DevicesV1alpha2Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := restclient.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &v1alpha2api.DevicesV1alpha2Client{client}, nil
}

func setConfigDefaults(config *restclient.Config) error {
	gv := v1alpha2.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme()).WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = restclient.DefaultKubernetesUserAgent()
	}

	return nil
}

func BuildConfigFromFlags(masterUrl, kubeconfigPath string) (*restclient.Config, error) {
	if kubeconfigPath == "" && masterUrl == "" {
		klog.Warning("Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.")
		kubeconfig, err := restclient.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}
		klog.Warning("error creating inClusterConfig, falling back to default config: ", err)
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterUrl}}).ClientConfig()
}

func InitCrdClient() (*v1alpha2api.DevicesV1alpha2Client, error) {
	ConfigPath := "./config.yaml"
	connectionConfig, err := ParseConfig(ConfigPath)
	if err != nil {
		print("parse config failed.")
		os.Exit(-1)
	}
	crdClient := InitRestClient(connectionConfig.KubeAPIConfig)
	return crdClient, err
}
