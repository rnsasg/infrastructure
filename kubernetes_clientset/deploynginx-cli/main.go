package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var (
	versionFlag   = flag.String("version", "", "Version of nginx to deploy")
	scaleFlag     = flag.Int("scale", 0, "Number of replicas to scale the deployment to")
	createFlag    = flag.Bool("create", false, "Create nginx deployment")
	nginxReplicas = flag.Int("replicas", 1, "Number of replicas for nginx deployment")
	kubeconfig    = flag.String("kubeconfig", "", "Path to the kubeconfig file")
)

func main() {
	flag.Parse()

	config, err := buildConfig(*kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubernetes clientset: %v\n", err)
		os.Exit(1)
	}

	if *createFlag {
		err := createDeployment(clientset, *nginxReplicas)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating nginx deployment: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("nginx deployment created with %d replicas\n", *nginxReplicas)
	}

	if *versionFlag != "" {
		err := upgradeDeployment(clientset, *versionFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error upgrading deployment: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deployment upgraded to version %s\n", *versionFlag)
	}

	if *scaleFlag > 0 {
		err := scaleDeployment(clientset, *scaleFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scaling deployment: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deployment scaled to %d replicas\n", *scaleFlag)
	}
}

func buildConfig(kubeconfigPath string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func createDeployment(clientset *kubernetes.Clientset, replicas int) error {
	deploymentClient := clientset.AppsV1().Deployments("default")

	nginxDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := deploymentClient.Create(context.TODO(), nginxDeployment, metav1.CreateOptions{})
	return err
}

func upgradeDeployment(clientset *kubernetes.Clientset, version string) error {
	deploymentClient := clientset.AppsV1().Deployments("default")

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ctx := context.TODO()
		deployment, getErr := deploymentClient.Get(ctx, "nginx-deployment", metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get latest version of Deployment: %v", getErr)
		}

		deployment.Spec.Template.Spec.Containers[0].Image = "nginx:" + version

		_, updateErr := deploymentClient.Update(ctx, deployment, metav1.UpdateOptions{})
		return updateErr
	})

	return retryErr
}

func scaleDeployment(clientset *kubernetes.Clientset, replicas int) error {
	deploymentClient := clientset.AppsV1().Deployments("default")

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ctx := context.TODO()
		deployment, getErr := deploymentClient.Get(ctx, "nginx-deployment", metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get latest version of Deployment: %v", getErr)
		}

		deployment.Spec.Replicas = int32Ptr(int32(replicas))

		_, updateErr := deploymentClient.Update(ctx, deployment, metav1.UpdateOptions{})
		return updateErr
	})

	return retryErr
}

func int32Ptr(i int32) *int32 { return &i }
