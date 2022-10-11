# kubectl-risky-webhooks

Check a cluster for potentially risky webhooks

## Example Usage

```sh
> kubectl-risky-webhooks

Checking for risky webhooks...

  NAME                                                                      WEBHOOK                     REPLICAS   HAS PDB   KUBE-SYSTEM IGNORED   IGNORES FAILURE  
 ⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤ ⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤ ⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤ ⏤⏤⏤⏤⏤⏤⏤⏤⏤ ⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤ ⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤⏤ 
  ValidatingWebhookConfigurations/kyverno-resource-validating-webhook-cfg   validate.kyverno.svc-fail          3   Yes       No                    No               

```
