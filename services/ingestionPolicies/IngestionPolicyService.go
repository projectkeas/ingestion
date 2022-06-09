package ingestionPolicies

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	spec "github.com/cloudevents/sdk-go/v2/binding/spec"
	types "github.com/projectkeas/crds/pkg/apis/keas.io/v1alpha1"
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

var (
	specs = spec.New().Version("1.0")
)

type IngestionPolicyService interface {
	GetDecision(event cloudevents.Event, data map[string]interface{}) (IngestionPolicyDecision, error)
}

type ingestionExecutionService struct {
	opa      *opa.OPAService
	versions map[string]string
}

func (ies *ingestionExecutionService) GetDecision(event cloudevents.Event, data map[string]interface{}) (IngestionPolicyDecision, error) {
	result := &IngestionPolicyDecision{
		Allow: true,
	}

	keys := ies.opa.GetPolicyKeys()
	if len(keys) == 0 {
		return *result, nil
	}

	subject := map[string]interface{}{
		"metadata": map[string]interface{}{
			specs.AttributeFromKind(spec.DataContentType).Name(): event.DataContentType(),
			specs.AttributeFromKind(spec.DataSchema).Name():      event.DataSchema(),
			specs.AttributeFromKind(spec.ID).Name():              event.ID(),
			specs.AttributeFromKind(spec.Source).Name():          event.Source(),
			specs.AttributeFromKind(spec.SpecVersion).Name():     event.SpecVersion(),
			specs.AttributeFromKind(spec.Subject).Name():         event.Subject(),
			specs.AttributeFromKind(spec.Time).Name():            event.Time(),
			specs.AttributeFromKind(spec.Type).Name():            event.Type(),
		},
		"payload": data,
	}

	for _, key := range keys {
		decision, err := ies.opa.EvaluatePolicy(key, subject)

		if err != nil {
			return IngestionPolicyDecision{}, err
		}

		allow := decision[0].Bindings["allow"].(bool)
		if !allow {
			result.Allow = false
			return *result, nil
		}
	}

	return *result, nil
}

func New() IngestionPolicyService {

	opa := &opa.OPAService{}
	svc := &ingestionExecutionService{
		opa:      opa,
		versions: map[string]string{},
	}

	informer := services.GetInformer()
	ingestionPoliciesFactory := informer.Keas().V1alpha1().IngestionPolicies()
	ingestionPoliciesFactory.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNewIngestionPolicy(svc),
		UpdateFunc: onUpdatedIngestionPolicy(svc),
		DeleteFunc: onDeletedIngestionPolicy(svc),
	}, 2*time.Minute)

	informer.Start(wait.NeverStop)
	informer.WaitForCacheSync(wait.NeverStop)

	return svc
}

func onNewIngestionPolicy(svc *ingestionExecutionService) func(policyInterface interface{}) {
	return func(policyInterface interface{}) {
		ingestionPolicy, successfulCast := policyInterface.(*types.IngestionPolicy)
		if successfulCast {
			if addOrUpdateIngestionPolicy(svc, ingestionPolicy) {
				log.Logger.Info("added new ingestion policy.", zap.Any("ingestionPolicy", map[string]string{
					"name":      ingestionPolicy.Name,
					"namespace": ingestionPolicy.Namespace,
				}))
			}
		} else {
			log.Logger.Error("could not cast ingestion policy")
		}
	}
}

func onUpdatedIngestionPolicy(svc *ingestionExecutionService) func(oldPolicyInterface interface{}, newPolicyInterface interface{}) {
	return func(oldPolicyInterface interface{}, newPolicyInterface interface{}) {
		ingestionPolicy, successfulCast := newPolicyInterface.(*types.IngestionPolicy)
		if successfulCast {
			if addOrUpdateIngestionPolicy(svc, ingestionPolicy) {
				log.Logger.Info("updated ingestion policy.", zap.Any("ingestionPolicy", map[string]string{
					"name":      ingestionPolicy.Name,
					"namespace": ingestionPolicy.Namespace,
				}))
			}
		} else {
			log.Logger.Error("could not cast ingestion policy")
		}
	}
}

func addOrUpdateIngestionPolicy(svc *ingestionExecutionService, ingestionPolicy *types.IngestionPolicy) bool {
	version, found := svc.versions[ingestionPolicy.Name]
	if found && version == ingestionPolicy.ResourceVersion {
		return false
	}

	allow := false
	allow = ingestionPolicy.Spec.Defaults.Allow

	svc.opa.AddOrUpdatePolicy("keas.ingestion", ingestionPolicy.Name, map[string]interface{}{
		"allow": allow,
	}, ingestionPolicy.Spec.Policy)
	svc.versions[ingestionPolicy.Name] = ingestionPolicy.ResourceVersion
	return true
}

func onDeletedIngestionPolicy(svc *ingestionExecutionService) func(policyInterface interface{}) {
	return func(policyInterface interface{}) {
		ingestionPolicy, successfulCast := policyInterface.(*types.IngestionPolicy)
		if successfulCast {
			svc.opa.RemovePolicy(ingestionPolicy.Namespace, ingestionPolicy.Name)
			delete(svc.versions, ingestionPolicy.Name)
			log.Logger.Info("Deleted ingestion policy. the policy is no longer in effect", zap.Any("ingestionPolicy", map[string]string{
				"name":      ingestionPolicy.Name,
				"namespace": ingestionPolicy.Namespace,
			}))
		} else {
			log.Logger.Error("could not cast ingestion policy")
		}
	}
}
