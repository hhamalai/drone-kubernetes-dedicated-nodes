#!/usr/bin/env bash
#
# Provision self-signed TLS certificate and key, store them in AWS SecretsManager.
# Deploy drone pod admission controller and webhook using the certificate.
#
set -eu

export IMAGE=$1

openssl genrsa -out webhookCA.key 2048
openssl req -new -key ./webhookCA.key -subj "/CN=drone-pod-admission.drone.svc" -out ./webhookCA.csr
openssl x509 -req -days 3650 -in webhookCA.csr -signkey webhookCA.key -out webhook.crt
export CA_BUNDLE=$(base64 ./webhook.crt)
echo '{}' | jq --arg key "$(<./webhookCA.key)"  --arg cert "$(<./webhook.crt)" '{"key.pem": $key, "cert.pem": $cert}' > cert-secrets.json

aws secretsmanager put-secret-value --secret-id drone-pod-admission-certs --secret-string file://cert-secrets.json

envsubst < ./webhook-deployment.yml | kubectl apply -f -
envsubst < ./webhook-registration.yml | kubectl apply -f -