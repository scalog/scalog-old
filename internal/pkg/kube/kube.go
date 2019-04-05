package kube

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

/*
InitKubernetesClient returns a client for interacting with the
kubernetes API server.
*/
func InitKubernetesClient() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

/*
AllPodsAreRunning returns true if and only if [pods] contains
only kubernetes pod objects that are in the "running" state.
*/
func AllPodsAreRunning(pods []v1.Pod) bool {
	for _, pod := range pods {
		if pod.Status.Phase != v1.PodRunning {
			return false
		}
	}
	return true
}

/*
GetShardPods blocks until the specified number of pods have appeared
*/
func GetShardPods(clientset *kubernetes.Clientset, query metav1.ListOptions, expectedPodCount int, namespace string) (*v1.PodList, error) {
	for {
		pods, err := clientset.CoreV1().Pods(namespace).List(query)
		if err != nil {
			return nil, err
		}
		if len(pods.Items) == expectedPodCount && AllPodsAreRunning(pods.Items) {
			return pods, nil
		}
		//wait for pods to start up
		time.Sleep(1000 * time.Millisecond)
	}
}
