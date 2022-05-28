package ingestionPolicies

import (
	"encoding/json"
	"time"

	types "github.com/projectkeas/crds/pkg/apis/keas.io/v1alpha1"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/ingestion/services"
	log "github.com/projectkeas/sdks-service/logger"
	"github.com/projectkeas/sdks-service/opa"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

const (
	SERVICE_NAME string = "IngestionPolicies"
)

type IngestionPolicyService interface {
	GetDecision(event sdk.EventEnvelope) (IngestionPolicyDecision, error)
}

type ingestionExecutionService struct {
	opa *opa.OPAService
}

func (ies *ingestionExecutionService) GetDecision(event sdk.EventEnvelope) (IngestionPolicyDecision, error) {
	result := &IngestionPolicyDecision{
		Allow: true,
		TTL:   -1,
	}

	keys := ies.opa.GetPolicyKeys()
	if len(keys) == 0 {
		return *result, nil
	}

	for _, key := range keys {
		decision, err := ies.opa.EvaluatePolicy(key, event)

		if err != nil {
			return IngestionPolicyDecision{}, err
		}

		allow := decision[0].Bindings["allow"].(bool)
		ttl, err := decision[0].Bindings["ttl"].(json.Number).Int64()
		if err != nil {
			return IngestionPolicyDecision{}, err
		}

		if !allow {
			result.Allow = false
		}

		if ttl > result.TTL && ttl != 0 {
			result.TTL = ttl
		}
	}

	return *result, nil
}

func New() IngestionPolicyService {

	opa := &opa.OPAService{}

	informer := services.GetInformer()
	ingestionPoliciesFactory := informer.Keas().V1alpha1().IngestionPolicies()
	ingestionPoliciesFactory.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNewIngestionPolicy(opa),
		UpdateFunc: onUpdatedIngestionPolicy(opa),
		DeleteFunc: onDeletedIngestionPolicy(opa),
	}, 2*time.Minute)

	informer.Start(wait.NeverStop)
	informer.WaitForCacheSync(wait.NeverStop)

	return &ingestionExecutionService{
		opa: opa,
	}
}

func onNewIngestionPolicy(opa *opa.OPAService) func(policyInterface interface{}) {
	return func(policyInterface interface{}) {
		ingestionPolicy, successfulCast := policyInterface.(*types.IngestionPolicy)
		if successfulCast {
			addOrUpdateIngestionPolicy(opa, ingestionPolicy)
			log.Logger.Info("added new ingestion policy.", zap.Any("ingestionPolicy", map[string]string{
				"name":      ingestionPolicy.Name,
				"namespace": ingestionPolicy.Namespace,
			}))
		} else {
			log.Logger.Error("could not cast ingestion policy")
		}
	}
}

func onUpdatedIngestionPolicy(opa *opa.OPAService) func(oldPolicyInterface interface{}, newPolicyInterface interface{}) {
	return func(oldPolicyInterface interface{}, newPolicyInterface interface{}) {
		ingestionPolicy, successfulCast := newPolicyInterface.(*types.IngestionPolicy)
		if successfulCast {
			addOrUpdateIngestionPolicy(opa, ingestionPolicy)
			log.Logger.Info("updated ingestion policy.", zap.Any("ingestionPolicy", map[string]string{
				"name":      ingestionPolicy.Name,
				"namespace": ingestionPolicy.Namespace,
			}))
		} else {
			log.Logger.Error("could not cast ingestion policy")
		}
	}
}

func addOrUpdateIngestionPolicy(opa *opa.OPAService, ingestionPolicy *types.IngestionPolicy) {
	allow := false
	ttl := -1

	if ingestionPolicy.Spec.Defaults.Allow == true {
		allow = true
	}

	if ingestionPolicy.Spec.Defaults.TTL != 0 {
		ttl = ingestionPolicy.Spec.Defaults.TTL
	}

	opa.AddOrUpdatePolicy("keas.ingestion", ingestionPolicy.Name, map[string]interface{}{
		"allow": allow,
		"ttl":   ttl,
	}, ingestionPolicy.Spec.Policy)
}

func onDeletedIngestionPolicy(opa *opa.OPAService) func(policyInterface interface{}) {
	return func(policyInterface interface{}) {
		ingestionPolicy, successfulCast := policyInterface.(*types.IngestionPolicy)
		if successfulCast {
			opa.RemovePolicy(ingestionPolicy.Namespace, ingestionPolicy.Name)
			log.Logger.Info("deleted ingestion policy. the policy is no longer in effect", zap.Any("ingestionPolicy", map[string]string{
				"name":      ingestionPolicy.Name,
				"namespace": ingestionPolicy.Namespace,
			}))
		} else {
			log.Logger.Error("could not cast ingestion policy")
		}
	}
}
