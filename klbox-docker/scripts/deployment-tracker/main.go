package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func connectConfig() (*rest.Config, error) {
	host, ok := os.LookupEnv("KUBERNETES_HOST")
	if !ok {
		return rest.InClusterConfig()
	}

	return &rest.Config{Host: host}, nil
}

func kubeClient() (*kubernetes.Clientset, error) {
	c, err := connectConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(c)
}

func checkDeploymentReady(ctx context.Context, k *kubernetes.Clientset, namespace, name string) (bool, error) {
	d, err := k.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	for _, c := range d.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionTrue {
			return true, nil
		}
	}
	return false, nil
}

func main() {
	var output string
	flag.StringVar(&output, "output", "/tmp/kl/deployment-tracker/status.json", "--output")
	flag.Parse()

	const (
		AgentDeployment         string = "kl-agent"
		AgentNamespace          string = "kloudlite"
		AgentOperatorDeployment string = "kl-agent-operator"

		GatewayNamespace  string = "kl-gateway"
		GatewayDeployment string = "default"
	)

	if err := os.MkdirAll(filepath.Dir(output), 0o777); err != nil {
		panic(err)
	}

	f, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	k, err := kubeClient()
	if err != nil {
		panic(err)
	}

	for {
		start := time.Now()
		agent, err := checkDeploymentReady(context.TODO(), k, AgentNamespace, AgentDeployment)
		if err != nil {
			slog.Error("failed to check deployment status", slog.Group("deployment", "namespace", AgentNamespace, "name", AgentDeployment))
		}

		agentOp, err := checkDeploymentReady(context.TODO(), k, AgentNamespace, AgentOperatorDeployment)
		if err != nil {
			slog.Error("failed to check deployment status", slog.Group("deployment", "namespace", AgentNamespace, "name", AgentOperatorDeployment))
		}

		gateway, err := checkDeploymentReady(context.TODO(), k, GatewayNamespace, GatewayDeployment)
		if err != nil {
			slog.Error("failed to check deployment status", slog.Group("deployment", "namespace", GatewayNamespace, "name", GatewayDeployment))
		}

		b, err := json.Marshal(map[string]any{
			"lastCheckedAt": start.Format(time.RFC3339),
			"compute":       agent && agentOp,
			"gateway":       gateway,
		})
		if err != nil {
			slog.Error("failed to marshal status data, got", "err", err)
		}

		if _, err := f.WriteAt(b, 0); err != nil {
			slog.Error("failed to persist status, got", "err", err)
		}

		slog.Info("written status", "file", f.Name())
		<-time.After(1 * time.Second)
	}
}
