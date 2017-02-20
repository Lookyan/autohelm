package kubeutils

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/api/unversioned"
)

func GetDeploymentPod(
	clientset *kubernetes.Clientset,
	namespace string,
	deploymentName string,
	containerName string) (string, error) {

	deployment, err := clientset.ExtensionsV1beta1Client.Deployments(namespace).
		Get(deploymentName)

	if err != nil {
		fmt.Printf("Can't attach to %s -> %s\nDeployment not found", deploymentName, containerName)
		return "", err
	}

	selector, err := unversioned.LabelSelectorAsSelector(deployment.Spec.Selector)

	if err != nil {
		return "", err
	}

	pods, err := clientset.CoreV1Client.Pods(namespace).
		List(v1.ListOptions{LabelSelector: selector.String()})

	if err != nil {
		return "", err
	}

	if len(pods.Items) > 0 {
		return pods.Items[0].GetName(), nil
	}

	return "", nil
}
