package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	reset   = "\033[0m"
	yellow  = "\033[33m"
	cyan    = "\033[36m"
	magenta = "\033[35m"
)

func GetK8sConfigMap(k8sConfigMap *map[string]string) error {
	// 构建 kubeconfig 文件路径
	//var kubeconfig string
	//if home := homedir.HomeDir(); home != "" {
	//	kubeconfig = filepath.Join(home, ".kube", "config")
	//}
	kubeconfig := "/path/to/your/kubeconfig"

	// 使用 kubeconfig 创建配置
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("无法构建配置: %v", err)
	}

	// 创建一个新的客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("无法创建客户端: %v", err)
		return err
	}

	// 获取命名空间为 "default" 的名为 "example-configmap" 的 ConfigMap
	namespace := "default"
	configMapName := "example-configmap"
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("无法获取 ConfigMap: %v", err)
		return err
	}
	*k8sConfigMap = configMap.Data
	return nil
}

func GetApollo(configMap *map[string]string) error {
	url := "http://example-apollo-config-server/configfiles/json/YourProject/default/your-configmap"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = json.Unmarshal(body, &configMap)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func compareMaps(map1, map2 map[string]string) {
	// 检查 apollo 中有但 k8s configMap 中没有的键
	fmt.Println("Apollo 中存在但在 Kubernetes ConfigMap 中不存在")
	for key, value := range map1 {
		if _, exists := map2[key]; !exists {
			fmt.Fprintf(os.Stdout, "Key: '%s%s%s'\n Apollo: '%s%s%s'\n", yellow, key, reset, yellow, value, reset)
		}
	}
	fmt.Println("Apollo 与 Kubernetes ConfigMap 值不同")
	for key, value := range map1 {
		if key == "CHAT_QUEST_GEN_PROMPT_DEFAULT" {
			continue
		}
		if val, exists := map2[key]; !exists {
		} else if val != value {
			fmt.Fprintf(os.Stdout, "Key: '%s%s%s' \n Apollo: '%s%s%s' \n Config: '%s%s%s'\n", yellow, key, reset, yellow, value, reset, yellow, val, reset)
		}
	}
}
func main() {
	configMap := make(map[string]string)
	k8sConfigMap := make(map[string]string)
	err := GetApollo(&configMap)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = GetK8sConfigMap(&k8sConfigMap)
	if err != nil {
		fmt.Println(err)
		return
	}
	compareMaps(configMap, k8sConfigMap)
}
