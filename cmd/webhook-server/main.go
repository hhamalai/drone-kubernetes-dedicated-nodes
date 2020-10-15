package main

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"

	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	mutatingwh "github.com/slok/kubewebhook/pkg/webhook/mutating"
)

func dronePodMutator(_ context.Context, obj metav1.Object) (bool, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		// If not a pod just continue the mutation chain(if there is one) and don't do nothing.
		return false, nil
	}

	pod.Spec.Tolerations = []corev1.Toleration{
		{
			Key:      "dedicated",
			Operator: "Equal",
			Value:    "CI",
			Effect:   "NoExecute",
		},
	}

	pod.Spec.NodeSelector = map[string]string{
		"dedicated": "CI",
	}

	var userId int64 = 1000
	var groupId int64 = 1000

	forceNoRoot := pod.Annotations["force-no-root"]
	if forceNoRoot == "true" {
		for i := range pod.Spec.Containers {
			if pod.Spec.Containers[i].SecurityContext == nil {
				pod.Spec.Containers[i].SecurityContext = &corev1.SecurityContext{}
			}

			if pod.Spec.Containers[i].SecurityContext.Privileged == nil || !*pod.Spec.Containers[i].SecurityContext.Privileged {
				pod.Spec.Containers[i].SecurityContext.RunAsUser = &userId
				pod.Spec.Containers[i].SecurityContext.RunAsGroup = &groupId
			}
		}
	}

	pod.Annotations["mutated"] = "true"
	pod.Annotations["mutator"] = "drone-pod-mutator"

	return false, nil
}


type config struct {
	certFile 			string
	keyFile  			string
}


func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", 	os.Getenv("TLS_CERT_FILE"), "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", os.Getenv("TLS_KEY_FILE"), "TLS key file")

	err := fl.Parse(os.Args[1:])
	if err != nil || cfg.certFile == "" || cfg.keyFile == "" {
		fl.Usage()
		return nil
	}
	return cfg
}

func main() {
	logger := &log.Std{Debug: true}

	cfg := initFlags()
	if cfg == nil {
		return
	}

	mt := mutatingwh.MutatorFunc(dronePodMutator)

	mcfg := mutatingwh.WebhookConfig{
		Name: "dronePod",
		Obj:  &corev1.Pod{},
	}
	wh, err := mutatingwh.NewWebhook(mcfg, mt, nil, nil, logger)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}

	// Get the handler for our webhook.
	whHandler, err := whhttp.HandlerFor(wh)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating webhook handler: %s", err)
		os.Exit(1)
	}
	logger.Infof("Listening on :8080")
	err = http.ListenAndServeTLS(":8080", cfg.certFile, cfg.keyFile, whHandler)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}