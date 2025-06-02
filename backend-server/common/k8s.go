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
	k8sInitonce sync.Once
	k8sClient   *kubernetes.Clientset
	config      *rest.Config
)

func initK8sClientSet() (*kubernetes.Clientset, *rest.Config) {
	k8sInitonce.Do(func() {
		config = ctrl.GetConfigOrDie()
		k8sClient = kubernetes.NewForConfigOrDie(config)
		klog.Infoln("init k8s clients")
	})

	return k8sClient, config
}

func GetPvcAnnotation(bflUser string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	initK8sClientSet()
	key := "userspace_pvc"
	namespace := "user-space-" + bflUser

	bfl, err := k8sClient.AppsV1().StatefulSets(namespace).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		klog.Errorln("find user's bfl error, ", err, ", ", namespace)
		return "", err
	}

	klog.Infof("bfl.Annotations: %+v", bfl.Annotations)
	klog.Infof("bfl.Annotations[%s]: %s at time %s", key, bfl.Annotations[key], time.Now().Format(time.RFC3339))
	return bfl.Annotations[key], nil
}
