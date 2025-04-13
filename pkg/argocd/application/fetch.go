package application

import (
	"argocd-watcher/pkg/argocd"
	"context"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var InstanceLabel string = "argocd.argoproj.io/instance"

func getApplications() ([]v1alpha1.Application, error) {
	argoClient := argocd.GetArgoCDClient()
	apps, err := argoClient.ArgoprojV1alpha1().Applications(argocd.ArgocdNs).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return apps.Items, nil
}

func getApplication(name string) (*v1alpha1.Application, error) {
	argoClient := argocd.GetArgoCDClient()
	return argoClient.ArgoprojV1alpha1().Applications(argocd.ArgocdNs).Get(context.TODO(), name, metav1.GetOptions{})
}

func GetApplicationTrack(name string) []*v1alpha1.Application {
	track := []*v1alpha1.Application{}
	for i := 0; i < 10; i++ {
		app, err := StoreGet(name)
		if err != nil {
			break
		}
		track = append(track, &app)
		if instance, ok := app.Labels[InstanceLabel]; ok {
			name = instance
		} else {
			break
		}
	}
	return track
}
