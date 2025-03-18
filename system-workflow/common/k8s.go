package common

import (
	"context"
	"time"

	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	once      sync.Once
	k8sClient *kubernetes.Clientset
	config    *rest.Config
)

func GetClientSet() (*kubernetes.Clientset, *rest.Config) {
	once.Do(func() {
		config = ctrl.GetConfigOrDie()
		k8sClient = kubernetes.NewForConfigOrDie(config)
		klog.Infoln("init k8s clients")
	})

	return k8sClient, config
}

func K8sTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	namespaces, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorln("Failed to list namespaces: ", err)
	}

	for _, ns := range namespaces.Items {
		_, err := k8sClient.AppsV1().StatefulSets(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to list StatefulSets in namespace %s: %v", ns.Name, err)
			continue
		}
		klog.Infoln("userspace: ", ns.Name, "bfl_name: ", ns.Name[len("user-space-"):], "at time: ", time.Now().Format(time.RFC3339))

	}

}
