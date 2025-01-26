package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

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
	err = waitForPod(pod)
	if err != nil {
		fmt.Println(err);
		return nil;
	}
	fmt.Println("Running Terminal")
	request := clientset.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command: []string{"/bin/sh", "-c", "stty -echo; exec /bin/sh"},
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
		err = exec.StreamWithContext(context.TODO(),remotecommand.StreamOptions{
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

func waitForPod(pod v1.Pod) error {
	for {
		if pod.Status.Phase == v1.PodRunning {
			fmt.Println("Pod Started Running")
			return nil
		}else if pod.Status.Phase != v1.PodPending{
			return fmt.Errorf("pod %s failed to run",pod.Name);
		}
		time.Sleep(2 * time.Second)
	}
}
