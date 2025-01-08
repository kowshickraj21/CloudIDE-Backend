package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetK8SClient () (*kubernetes.Clientset, string, string, error) {
    home, _ := os.UserHomeDir()
    ingress := os.Getenv("INGRESS_NAME")
    namespace := os.Getenv("K8S_NAMESPACE")
	kubeConfigPath := filepath.Join(home, ".kube/config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil,ingress,namespace, err
	}
	client := kubernetes.NewForConfigOrDie(config)
    return client,ingress,namespace,nil
}

func StartDeployment(client *kubernetes.Clientset, ingressName,namespace string, details Stash) error {

    // hostPathType := corev1.HostPathDirectory
	var replicas int32 = 1
	deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name: details.Name,
            Annotations: map[string]string{
                "isRunning":"True",
                "LastOpened": time.Now().Format(time.RFC3339),
            },
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": details.Name,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": details.Name,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  details.Name,
                            Image: details.Image,
                            TTY: true,
                            Stdin: true,
                            Ports: []corev1.ContainerPort{
                                {
                                    ContainerPort: details.Port,
                                },
                            },
                            // VolumeMounts: []corev1.VolumeMount{
                            //     {
                            //         Name: "home-mount",
                            //         MountPath: "/home",
                            //     },
                            // },
                        },
                    },
                    // Volumes: []corev1.Volume{
                    //     {
                    //         Name: "home-mount",
                    //         VolumeSource: corev1.VolumeSource{
                    //             HostPath: &corev1.HostPathVolumeSource{
                    //                 Path: fmt.Sprintf("/hostmnt/s3-bucket/stash/%s",details.Name),
                    //                 Type: &hostPathType,
                    //             },
                    //         },
                    //     },
                    // },
                },
            },
        },
    }

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: details.Name,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":details.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port: 80,
					TargetPort: intstr.FromInt32(details.Port),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	deploymentsClient := client.AppsV1().Deployments(namespace)
	newDeployment, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	servicesClient := client.CoreV1().Services(namespace)
	newService, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return  err
	}

	err = updateIngress(client,namespace,ingressName,"/out/",details.Name,80)

    if err != nil {
        return err
    }

	fmt.Printf("Deployment '%s' and Service '%s' are created!", newDeployment.Name, newService.Name)
	return nil
}


func updateIngress(client *kubernetes.Clientset, namespace, ingressName, path, serviceName string, servicePort int32) error {
    ingressClient := client.NetworkingV1().Ingresses(namespace)

    ingress, err := ingressClient.Get(context.TODO(), ingressName, metav1.GetOptions{})
    if err != nil {
        return fmt.Errorf("failed to get ingress: %v", err)
    }

    pathType := networkingv1.PathTypePrefix
    newPath := networkingv1.HTTPIngressPath{
        Path:     path,
        PathType: &pathType,
        Backend: networkingv1.IngressBackend{
            Service: &networkingv1.IngressServiceBackend{
                Name: serviceName,
                Port: networkingv1.ServiceBackendPort{
                    Number: servicePort,
                },
            },
        },
    }

    if len(ingress.Spec.Rules) > 0 {
        ingress.Spec.Rules[0].HTTP.Paths = append(ingress.Spec.Rules[0].HTTP.Paths, newPath)
    } else {
        return fmt.Errorf("no rules found in the existing ingress")
    }

    _, err = ingressClient.Update(context.TODO(), ingress, metav1.UpdateOptions{})
    if err != nil {
        return fmt.Errorf("failed to update ingress: %v", err)
    }
	fmt.Println("Added to ingress")
    return nil
}


func StartStash(details Stash) error {
    client,ingress,namespace,err := GetK8SClient()
    if err != nil {
        return err
    }
    if IsDeploymentRunning(client,namespace,details.Name) {
        return nil
    }
    
    err = StartDeployment(client,ingress,namespace,details)
    if err != nil {
        return err
    }
    return nil
}


func IsDeploymentRunning(client *kubernetes.Clientset, namespace,deploymentName string) bool {
    deployment,err := client.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
    if err != nil {
    return false
    }
    if deployment.Name == deploymentName {
        return true
    }
    return false
}

func CloseStash(name string) {
    client,_,namespace,err := GetK8SClient()
    if err != nil {
        fmt.Println(err)
    }
    deployment,err := client.AppsV1().Deployments(namespace).Get(context.TODO(),name,metav1.GetOptions{})
    if err != nil {
        fmt.Println(err)
    }
    deployment.ObjectMeta.Annotations["isRunning"] = "False"
    deployment.ObjectMeta.Annotations["LastOpened"] = time.Now().Format(time.RFC3339)
    
    client.AppsV1().Deployments(namespace).Update(context.TODO(),deployment,metav1.UpdateOptions{})
}