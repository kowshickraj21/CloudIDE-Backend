package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"main/ws"

	"net/http"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// startPod("check","nodejs")
	client := AWSInit()
	if client == nil {
		log.Fatalln("Initialization Error!")
	}
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
	ws.StartSocket(w,r,client)
		
	})
}

func startPod(name, image string) {
    home, _ := os.UserHomeDir()
	kubeConfigPath := filepath.Join(home, ".kube/config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err.Error())
	}

	client := kubernetes.NewForConfigOrDie(config)
	namespace := "default"

    podsClient := client.CoreV1().Pods(namespace)

    yamlFile, err := ioutil.ReadFile("/repl/pods.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error reading YAML file: %s", err.Error()))
	}

	pod := &v1.Pod{}
	decode := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer().Decode
	_, _, err = decode(yamlFile, nil, pod)
	if err != nil {
		panic(fmt.Sprintf("Error decoding YAML file: %s", err.Error()))
	}

	pod.Name = name
	if len(pod.Spec.Containers) > 0 {
		pod.Spec.Containers[0].Image = image
	} else {
		panic("No containers found in the Pod spec")
	}

    newPod, err := podsClient.Create(context.TODO(), pod, metav1.CreateOptions{})
    if err != nil {
        panic(err.Error())
    }
    fmt.Printf("Pod '%s' is created!", newPod.Name)
   
}