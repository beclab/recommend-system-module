package common

import (
	"context"
	"strings"
	"time"

	"sync"

	"go.uber.org/zap"
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

func initK8sClientSet() (*kubernetes.Clientset, *rest.Config) {
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
	Logger.Info("K8sTest start")
	initK8sClientSet()
	namespaces, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorln("Failed to list namespaces: ", err)
	}
	Logger.Info("userspace", zap.Int(" len:", len(namespaces.Items)))
	for _, ns := range namespaces.Items {
		_, err := k8sClient.AppsV1().StatefulSets(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to list StatefulSets in namespace %s: %v", ns.Name, err)
			continue
		}
		Logger.Info("userspace: ", zap.String("ns name:", ns.Name), zap.String("bfl_user:", ns.Name[len("user-space-"):]))

	}
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

func GetUserList() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Logger.Info("K8s get userlist start")
	initK8sClientSet()
	namespaces, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorln("Failed to list namespaces: ", err)
	}
	Logger.Info("userspace", zap.Int(" len:", len(namespaces.Items)))
	userList := make([]string, 0)
	for _, ns := range namespaces.Items {
		_, err := k8sClient.AppsV1().StatefulSets(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to list StatefulSets in namespace %s: %v", ns.Name, err)
			continue
		}
		//Logger.Info("userspace: ", zap.String("ns name:", ns.Name))
		if strings.HasPrefix(ns.Name, "user-space-") {
			user := ns.Name[len("user-space-"):]
			userList = append(userList, user)
			Logger.Info("userspace: ", zap.String("ns name:", ns.Name), zap.String("bfl_user:", user))
		}

	}

	return userList
}
