package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lensesio/tableprinter"
	webhookv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	client  *kubernetes.Clientset
	ctx     = context.Background()
	showAll = false
)

func main() {
	var err error

	flag.BoolVar(&showAll, "show-all", false, "Show all webhooks")
	flag.Parse()

	client, err = getClient()
	if err != nil {
		log.Fatal("failed to get kubernetes client - ", err)
	}

	fmt.Printf("Checking for risky webhooks...\n\n")

	validatingWebhooks, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	mutatingWebhooks, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	webhooks := []Webhook{}

	for _, v := range validatingWebhooks.Items {
		for _, w := range v.Webhooks {
			if isPodWebhook(w.Rules) {
				replicas, err := getReplicas(w.ClientConfig)
				if err != nil {
					fmt.Printf("⚠️  Failed to check %s / %s - %v\n", v.ObjectMeta.Name, w.Name, err)
					continue
				}
				wh := Webhook{
					Name:              fmt.Sprintf("ValidatingWebhookConfigurations/%s", v.ObjectMeta.Name),
					Webhook:           w.Name,
					FailureIgnore:     *w.FailurePolicy == webhookv1.Ignore,
					Replicas:          replicas,
					PDB:               hasPDB(w.ClientConfig),
					KubeSystemIgnored: isKubeSystemIgnored(w.NamespaceSelector),
				}
				if isRisky(wh) || showAll {
					webhooks = append(webhooks, wh)
				}
			}
		}
	}
	for _, m := range mutatingWebhooks.Items {
		for _, w := range m.Webhooks {
			if isPodWebhook(w.Rules) {
				replicas, err := getReplicas(w.ClientConfig)
				if err != nil {
					fmt.Printf("⚠️  Failed to check %s / %s - %v\n", m.ObjectMeta.Name, w.Name, err)
					continue
				}
				wh := Webhook{
					Name:              fmt.Sprintf("MutatingWebhookConfiguration/%s", m.ObjectMeta.Name),
					Webhook:           w.Name,
					FailureIgnore:     *w.FailurePolicy == webhookv1.Ignore,
					Replicas:          replicas,
					PDB:               hasPDB(w.ClientConfig),
					KubeSystemIgnored: isKubeSystemIgnored(w.NamespaceSelector),
				}
				if isRisky(wh) || showAll {
					webhooks = append(webhooks, wh)
				}
			}
		}
	}

	if len(webhooks) > 0 {
		printer := tableprinter.New(os.Stdout)
		printer.RowSeparator = "⏤"
		printer.Print(webhooks)
	} else {
		fmt.Println("🎉 No risky webhooks found!")
	}
}

func getClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfigPath := os.Getenv("KUBECONFIG")
		if kubeconfigPath == "" {
			kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	return kubernetes.NewForConfig(config)
}

func isKubeSystemIgnored(namespaceSelector *v1.LabelSelector) bool {
	namespaceNameLabel := "kubernetes.io/metadata.name"
	if namespaceSelector != nil {
		if namespaceSelector.MatchLabels[namespaceNameLabel] == "kube-system" {
			return false
		}
		for _, expression := range namespaceSelector.MatchExpressions {
			if expression.Key == namespaceNameLabel && expression.Operator == v1.LabelSelectorOpNotIn {
				for _, val := range expression.Values {
					if strings.ToLower(val) == "kube-system" {
						return true
					}
				}
			}
		}
	}

	return false
}

func getReplicas(clientConfig webhookv1.WebhookClientConfig) (int32, error) {
	if clientConfig.Service != nil {

		svc, err := client.CoreV1().Services(clientConfig.Service.Namespace).Get(ctx, clientConfig.Service.Name, v1.GetOptions{})
		if err != nil {
			return 0, err
		}

		pods, err := client.CoreV1().Pods(clientConfig.Service.Namespace).List(ctx, v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(svc.Spec.Selector).String(),
		})
		if err != nil {
			return 0, err
		}

		return int32(len(pods.Items)), nil
	}

	return 0, nil
}

func hasPDB(clientConfig webhookv1.WebhookClientConfig) bool {
	if clientConfig.Service != nil {

		svc, err := client.CoreV1().Services(clientConfig.Service.Namespace).Get(ctx, clientConfig.Service.Name, v1.GetOptions{})
		if err != nil {
			panic(err)
		}

		pdbs, err := client.PolicyV1().PodDisruptionBudgets(clientConfig.Service.Namespace).List(ctx, v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(svc.Spec.Selector).String(),
		})
		if err != nil {
			panic(err)
		}
		return len(pdbs.Items) > 0
	}

	return false
}

func isRisky(webhook Webhook) bool {
	if webhook.FailureIgnore == true {
		return false
	}
	if webhook.PDB == false || webhook.Replicas < 2 {
		return true
	}
	return webhook.KubeSystemIgnored == false
}

func isPodWebhook(rules []webhookv1.RuleWithOperations) bool {
	for _, r := range rules {
		for _, resource := range r.Resources {
			if resource == "*" || strings.Contains(resource, "pods") {
				return true
			}
		}
	}

	return false
}
