package kubeadm

import (
	"fmt"
	"os"
	
	"github.com/kraken-pedestal/internal/config"
	"gopkg.in/yaml.v3"
)

// KubeadmInitYAML 对应 kubeadm InitConfiguration + ClusterConfiguration + KubeletConfiguration
type KubeadmInitYAML struct {
	APIVersion           string                `yaml:"apiVersion"`
	Kind                 string                `yaml:"kind"`
	LocalAPIEndpoint     *LocalAPIEndpoint     `yaml:"localAPIEndpoint"`
	NodeRegistration     *NodeRegistration     `yaml:"nodeRegistration"`
	ClusterConfiguration *ClusterConfiguration `yaml:"clusterConfiguration"`
}

type ClusterConfiguration struct {
	APIVersion           string           `yaml:"apiVersion"`
	Kind                 string           `yaml:"kind"`
	KubernetesVersion    string           `yaml:"kubernetesVersion"`
	ControlPlaneEndpoint string           `yaml:"controlPlaneEndpoint,omitempty"`
	Networking           Networking       `yaml:"networking"`
	APIServer            *APIServer       `yaml:"apiServer,omitempty"`
	ControllerManager    *ComponentConfig `yaml:"controllerManager,omitempty"`
	Scheduler            *ComponentConfig `yaml:"scheduler,omitempty"`
	Etcd                 *EtcdConfig      `yaml:"etcd,omitempty"`
	DNSType              string           `yaml:"dns,omitempty"`
}
type Networking struct {
	PodSubent     string `yaml:"podSubent,omitempty"`
	ServiceSubent string `yaml:"serviceSubent,omitempty"`
	DNSType       string `yaml:"dns,omitempty"`
}
type APIServer struct {
	ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
}
type ComponentConfig struct {
	ExtraArgs map[string]string `yaml:"extraArgs,omitempty"`
}
type EtcdConfig struct {
	Local *EtcdLocal `yaml:"local,omitempty"`
}
type EtcdLocal struct {
	DataDir string `yaml:"dataDir,omitempty"`
	Image   string `yaml:"image,omitempty"`
}

type InitConfiguration struct {
	APIVersion       string           `yaml:"apiVersion"`
	Kind             string           `yaml:"kind"`
	LocalAPIEndpoint LocalAPIEndpoint `yaml:"localAPIEndpoint,omitempty"`
	NodeRegistration NodeRegistration `yaml:"nodeRegistration,omitempty"`
	BootstrapTokens  []BootstrapToken `yaml:"bootstrapTokens,omitempty"`
}
type LocalAPIEndpoint struct {
	AdvertiseAddress string `yaml:"advertiseAddress,omitempty"`
	BindPort         int    `yaml:"bindPort,omitempty"`
}
type NodeRegistration struct {
	Name      string `yaml:"name,omitempty"`
	CRISocker string `yaml:"CRISocker,omitempty"`
}

type Taint struct {
	Key    string `yaml:"key"`
	Value  string `yaml:"value,omitempty"`
	Effect string `yaml:"effect"`
}
type BootstrapToken struct {
	Token string `yaml:"token,omitempty"`
}

type JoinConfiguration struct {
	APIVersion       string            `yaml:"apiVersion"`
	Kind             string            `yaml:"kind"`
	Discovery        Discovery         `yaml:"discovery"`
	NodeRegistration NodeRegistration  `yaml:"nodeRegistration,omitempty"`
	ControlPlane     *JoinControlPlane `yaml:"controlPlane,omitempty"`
}
type Discovery struct {
	BootstrapTokens   *BootstrapTokenDiscovery `yaml:"bootstrapTokens,omitempty"`
	TLSBootstrapToken string                   `yaml:"tlsBootstrapToken,omitempty"`
}
type BootstrapTokenDiscovery struct {
	Token             string   `yaml:"token"`
	APIServerEndpoint string   `yaml:"apiServerEndpoint"`
	CACertHashes      []string `yml:"caCertHashes,omitempty"`
}
type JoinControlPlane struct {
	LocalAPIEndpodint LocalAPIEndpoint `yaml:"localAPIEndpodint,omitempty"`
	CertificateKey    string           `yaml:"certificateKey,omitempty"`
}

// GenerateInitConfig 从，ClusterConfig生成 kubeadm init yaml 配置并写入文件
func (k *Kubeadm) GenerateInitConfig(clustercfg *config.ClusterConfig, nodeAddr string) error {
	// 1. 构建ClusterConfiguration
	clusterYAML := &ClusterConfiguration{
		APIVersion:           "Kubeadm.k8s.io/v1beta1",
		Kind:                 "ClusterConfiguration",
		KubernetesVersion:    clustercfg.KubernetesVersion,
		ControlPlaneEndpoint: clustercfg.ControlPlane.Endpoint,
		Networking: Networking{
			PodSubent:     clustercfg.PodCIDR,
			ServiceSubent: clustercfg.ServiceCIDR,
			DNSType:       clustercfg.ServiceDNSDomain,
		},
	}
	
	// 2. 从ExtraKubeadm 中获取 apiServer/controllerManager.scheduler/etcd 配置
	//
	// 【逻辑流程图】
	// ┌─────────────────────────────────────────────────────────────────────┐
	// │  clustercfg.ExtraKubeadm (map[string]interface{})                    │
	// │        │                                                            │
	// │        ▼                                                            │
	// │  ┌─────────────┐  取key["apiServer"]                               │
	// │  │  nil ?      │───否──→ 跳过整个逻辑块                            │
	// │  └─────────────┘                                                   │
	// │        │是                                                         │
	// │        ▼                                                            │
	// │  类型断言 .(map[string]interface{})                                  │
	// │        │                                                            │
	// │   ┌────┴────┐                                                       │
	// │   ▼         ▼                                                       │
	// │  成功      失败                                                      │
	// │   │         │                                                       │
	// │   ▼         └──→ 跳过                                               │
	// │  取 v["extraArgs"]                                                  │
	// │        │                                                            │
	// │   类型断言 .(map[string]interface{})                                 │
	// │        │                                                            │
	// │   ┌────┴────┐                                                       │
	// │   ▼         ▼                                                       │
	// │  成功      失败                                                      │
	// │   │         │                                                       │
	// │   ▼         └──→ 跳过                                               │
	// │  遍历 ea 中每个 key/value                                            │
	// │        │                                                            │
	// │        ▼                                                            │
	// │  fmt.Sprintf("%v", val) 转成 string                                 │
	// │        │                                                            │
	// │        ▼                                                            │
	// │  赋值给 clusterYAML.APIServer.ExtraArgs                              │
	// └─────────────────────────────────────────────────────────────────────┘
	//
	// 【为什么要三层判断？】
	// ExtraKubeadm 是 map[string]interface{}，Go 是静态类型语言，
	// 每次从 map 取值后都是 interface{}，必须逐级类型断言才能安全访问嵌套字段。
	// 任何一级类型不匹配就跳过，避免运行时 panic。
	//
	// 【类型转换说明】
	// map[string]interface{} ──遍历──→ interface{} 值 ──fmt.Sprintf──→ string
	// 最终目标: clusterYAML.APIServer.ExtraArgs 是 map[string]string
	if clustercfg.ExtraKubeadm != nil {
		
		//从 ExtraKubeadm 中取 apiServer 这个 key，并类型断言为 map[string]interface{}。
		//v 是取到的值
		//ok 是断言是否成功（如果 apiServer 不存在或类型不对，ok 为 false）
		
		if v, ok := clustercfg.ExtraKubeadm["apiServer"].(map[string]interface{}); ok {
			// 再从 apiServer 的map中取 extraArgs ，同样断言为map[string]interface{}
			if ea, ok := v["extraArgs"].(map[string]interface{}); ok {
				
				//创建一个 map[string]string，用来存放最终转换后的参数。注意：目标类型是 map[string]string，而源类型是 map[string]interface{}，所以需要转换。
				args := make(map[string]string)
				for k, val := range ea {
					//遍历 extraArgs 中的每一个键值对，用 fmt.Sprintf("%v", val) 把 interface{} 类型的值转换成字符串。
					args[k] = fmt.Sprintf("%v", val)
				}
				//最后把转换好的 args 赋值给 clusterYAML.APIServer。
				clusterYAML.APIServer = &APIServer{ExtraArgs: args}
			}
			
		}
	}
	
	// 3. 构建InitConfiguration
	initYAML := &InitConfiguration{
		APIVersion: "kubeadm.k8s.io/v1beta1",
		Kind:       "InitConfiguration",
		LocalAPIEndpoint: LocalAPIEndpoint{
			AdvertiseAddress: nodeAddr,
			BindPort:         6443,
		},
		NodeRegistration: NodeRegistration{
			CRISocker: clustercfg.ContainerRuntime.Socket,
		},
	}
	// 4. 合并写入: kubeadm 支持多文档yaml
	out, err := yaml.Marshal(initYAML)
	if err != nil {
		return fmt.Errorf("marshal InitConfiguration: %w", err)
	}
	bytesOutPut := append(out, []byte("\n---\n")...)
	marshal, err := yaml.Marshal(clusterYAML)
	if err != nil {
		return fmt.Errorf("marshal ClusterConfiguration: %w", err)
	}
	output := append(bytesOutPut, marshal...)
	
	// 5. 写入文件
	if err := os.WriteFile(k.configPath, output, 0664); err != nil {
		return fmt.Errorf("write kubeadm config to %s: %w", k.configPath, err)
	}
	return nil
}

// GenerateJoinConfig 从ClusterConfig 生成 kubeadm join yaml配置写入文件
func (k *Kubeadm) GenerateJoinConfig(clusterCfg *config.ClusterConfig, nodeAddr string, token string, apiServerEndpoint string, caCertHash string, isControlPlane bool) error {
	joinYAML := &JoinConfiguration{
		//TODO: 此处需要根据 kubeadm 版本动态设置
		APIVersion: "kubeadm.k8s.io/v1beta4",
		Kind:       "JoinConfiguration",
		Discovery: Discovery{
			BootstrapTokens: &BootstrapTokenDiscovery{
				Token:             token,
				APIServerEndpoint: apiServerEndpoint,
				CACertHashes:      []string{caCertHash},
			},
		},
		NodeRegistration: NodeRegistration{
			CRISocker: clusterCfg.ContainerRuntime.Socket,
		},
	}
	if isControlPlane {
		joinYAML.ControlPlane = &JoinControlPlane{
			LocalAPIEndpodint: LocalAPIEndpoint{
				AdvertiseAddress: nodeAddr,
				BindPort:         6443,
			},
			CertificateKey: clusterCfg.ControlPlane.CertificateKey,
		}
	}
	output, err := yaml.Marshal(joinYAML)
	if err != nil {
		return fmt.Errorf("marshal JoinConfiguration: %w", err)
	}
	if err := os.WriteFile(k.configPath, output, 0664); err != nil {
		return fmt.Errorf("write JoinConfiguration: %w", err)
	}
	return nil
}
