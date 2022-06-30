package cluster

import (
	"context"
	"fmt"
	apppkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	argoapp "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/arlonproj/arlon/pkg/common"
	logpkg "github.com/arlonproj/arlon/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

//------------------------------------------------------------------------------

func List(
	appIf argoapp.ApplicationServiceClient,
	config *restclient.Config,
	argocdNs string,
) (clist []Cluster, err error) {
	log := logpkg.GetLogger()
	// List legacy clusters (have clusterspec)
	apps, err := appIf.List(context.Background(),
		&apppkg.ApplicationQuery{Selector: "managed-by=arlon,arlon-type=cluster"})
	if err != nil {
		return nil, fmt.Errorf("failed to list argocd applications: %s", err)
	}
	for _, a := range apps.Items {
		clist = append(clist, Cluster{
			Name:            a.Name,
			ClusterSpecName: a.Annotations[common.ClusterSpecAnnotationKey],
			ProfileName:     a.Annotations[common.ProfileAnnotationKey],
		})
	}
	// List next-gen clusters (have base cluster)
	apps, err = appIf.List(context.Background(),
		&apppkg.ApplicationQuery{Selector: "managed-by=arlon,arlon-type=cluster-app"})
	if err != nil {
		return nil, fmt.Errorf("failed to list next-gen clusters: %s", err)
	}
	for _, a := range apps.Items {
		// Find any corresponding profile apps
		profileApps, err := appIf.List(context.Background(),
			&apppkg.ApplicationQuery{
				Selector: "managed-by=arlon,arlon-type=profile-app,arlon-cluster=" + a.Name,
			})
		if err != nil {
			return nil, fmt.Errorf(
				"failed to list profile apps associated with cluster %s: %s",
				a.Name, err,
			)
		}
		var profileName string
		if len(profileApps.Items) > 0 {
			// In theory, multiple profile apps can be associated with the same
			// cluster. For now, just report the first one found.
			profileName = profileApps.Items[0].Labels["arlon-profile"]
		}
		clist = append(clist, Cluster{
			Name:        a.Name,
			ProfileName: profileName,
			BaseCluster: &BaseClusterInfo{
				Name:         a.Annotations[baseClusterNameAnnotation],
				RepoUrl:      a.Annotations[baseClusterRepoUrlAnnotation],
				RepoRevision: a.Annotations[baseClusterRepoRevisionAnnotation],
				RepoPath:     a.Annotations[baseClusterRepoPathAnnotation],
			},
		})
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get kube client: %s", err)
	}
	corev1 := kubeClient.CoreV1()
	secrApi := corev1.Secrets(argocdNs)
	secrs, err := secrApi.List(context.Background(), metav1.ListOptions{
		LabelSelector: argoClusterSecretTypeLabel + "," + externalClusterTypeLabel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster secrets: %s", err)
	}
	for _, secr := range secrs.Items {
		clusterName := secr.Data["name"]
		if clusterName == nil {
			log.V(1).Info("cluster secret skipped because missing cluster name",
				"secretName", secr.Name)
		}
		clist = append(clist, Cluster{
			Name:        string(clusterName),
			ProfileName: secr.Annotations[common.ProfileAnnotationKey],
			IsExternal:  true,
			SecretName:  secr.Name,
		})
	}
	return
}
