package k8s

import (
	// "bufio"
	// "bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type TerminalSession struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
	Close  func()
}

func StartTerminal(deploymentName string) *TerminalSession {
	home, _ := os.UserHomeDir()
	kubeConfigPath := filepath.Join(home, ".kube/config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Println("Failed to build Kubernetes config:", err)
		return nil
	}

	clientset := kubernetes.NewForConfigOrDie(config)
	namespace := "default"

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentName),
	})
	if err != nil {
		fmt.Println("Error listing pods:", err)
		return nil
	}

	if len(podList.Items) == 0 {
		fmt.Println("No pods found for the deployment.")
		return nil
	}

	pod := podList.Items[0]
	fmt.Println("Running Terminal")
	request := clientset.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command: []string{"/bin/sh"},
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if err != nil {
		fmt.Println("Error setting up terminal exec:", err)
		return nil
	}

	stdinPipeReader, stdinPipeWriter := io.Pipe()
	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()

	go func() {
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  stdinPipeReader,
			Stdout: stdoutPipeWriter,
			Stderr: stderrPipeWriter,
			Tty:    true,
		})
		stdinPipeReader.Close()
		stdoutPipeWriter.Close()
		stderrPipeWriter.Close()
	}()

	return &TerminalSession{
		Stdin:  stdinPipeWriter,
		Stdout: stdoutPipeReader,
		Stderr: stderrPipeReader,
		Close: func() {
			stdinPipeWriter.Close()
			stdoutPipeReader.Close()
			stderrPipeReader.Close()
		},
	}
}
