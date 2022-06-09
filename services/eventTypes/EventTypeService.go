package eventTypes

import (
	"fmt"
	"time"

	"github.com/projectkeas/ingestion/services"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	types "github.com/projectkeas/crds/pkg/apis/keas.io/v1alpha1"
	log "github.com/projectkeas/sdks-service/logger"
	jsonSchema "github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	SERVICE_NAME string = "EventTypes"
)

type EventTypeService interface {
	Validate(event cloudevents.Event, data map[string]interface{}) error
}

type eventTypesExecutionService struct {
	eventTypes map[string]validatableEventType
}

func (service eventTypesExecutionService) Validate(event cloudevents.Event, data map[string]interface{}) error {

	key := event.DataSchema()

	if key == "" {
		// TODO :: use server context to decide whether to allow schemaless cloudevents
		return nil
	}

	vt, found := service.eventTypes[key]
	if found {
		return vt.Validate(data)
	}

	return fmt.Errorf("no matching schema found for: %s", key)
}

func New() EventTypeService {

	informerFactory := services.GetInformer()
	service := &eventTypesExecutionService{}

	eventTypesFactory := informerFactory.Keas().V1alpha1().EventTypes()
	eventTypesFactory.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    onNewEventType(service),
		UpdateFunc: onUpdatedEventType(service),
		DeleteFunc: onDeletedEventType(service),
	}, 2*time.Minute)

	// Ensure that our informers have been started and we have valid caches
	informerFactory.Start(wait.NeverStop)
	informerFactory.WaitForCacheSync(wait.NeverStop)

	return service
}

func onNewEventType(service *eventTypesExecutionService) func(eventTypeInterface interface{}) {
	return func(policyInterface interface{}) {
		eventType, successfulCast := policyInterface.(*types.EventType)
		if successfulCast {
			if addOrUpdateEventType(service, eventType) {
				log.Logger.Info("added new event type", zap.Any("eventType", map[string]string{
					"name":      eventType.Name,
					"namespace": eventType.Namespace,
					"schemaUri": eventType.Spec.SchemaUri,
				}))
			}
		} else {
			log.Logger.Error("could not cast event type")
		}
	}
}

func onUpdatedEventType(service *eventTypesExecutionService) func(oldEventTypeInterface interface{}, newEventTypeInterface interface{}) {
	return func(oldEventType interface{}, newEventType interface{}) {
		eventType, successfulCast := newEventType.(*types.EventType)
		if successfulCast {
			if addOrUpdateEventType(service, eventType) {
				log.Logger.Info("updated event type", zap.Any("eventType", map[string]string{
					"name":      eventType.Name,
					"namespace": eventType.Namespace,
					"schemaUri": eventType.Spec.SchemaUri,
				}))
			}
		} else {
			log.Logger.Error("could not cast event type")
		}
	}
}

func addOrUpdateEventType(service *eventTypesExecutionService, eventType *types.EventType) bool {
	if service.eventTypes == nil {
		service.eventTypes = map[string]validatableEventType{}
	}

	et, found := service.eventTypes[eventType.Spec.SchemaUri]
	if (found) && et.version == eventType.ResourceVersion {
		return false
	}

	schema, err := jsonSchema.CompileString("schema.json", eventType.Spec.Schema)
	if err != nil {
		log.Logger.Error("Cannot parse json schema. Not adding schema to collection", zap.Any("eventType", map[string]string{
			"name":      eventType.Name,
			"namespace": eventType.Namespace,
			"schemaUri": eventType.Spec.SchemaUri,
		}), zap.Error(err))
		return false
	}

	service.eventTypes[eventType.Spec.SchemaUri] = validatableEventType{
		schema:    *schema,
		schemaUri: eventType.Spec.SchemaUri,
		version:   eventType.ResourceVersion,
	}

	return true
}

func onDeletedEventType(service *eventTypesExecutionService) func(eventTypeInterface interface{}) {
	return func(policyInterface interface{}) {
		eventType, successfulCast := policyInterface.(*types.EventType)
		if successfulCast {
			delete(service.eventTypes, eventType.Spec.SchemaUri)

			log.Logger.Info("deleted event type", zap.Any("eventType", map[string]string{
				"name":      eventType.Name,
				"namespace": eventType.Namespace,
				"schemaUri": eventType.Spec.SchemaUri,
			}))
		} else {
			log.Logger.Error("could not cast event type")
		}
	}
}
