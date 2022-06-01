package eventTypes

import (
	"fmt"
	"time"

	types "github.com/projectkeas/crds/pkg/apis/keas.io/v1alpha1"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/ingestion/services"
	log "github.com/projectkeas/sdks-service/logger"
	jsonSchema "github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

const (
	SERVICE_NAME string = "EventTypes"
)

type EventTypeService interface {
	Validate(event sdk.EventEnvelope) error
}

type eventTypesExecutionService struct {
	eventTypes map[string]validatableEventType
}

func (service eventTypesExecutionService) Validate(event sdk.EventEnvelope) error {

	subType := "*"
	if event.Metadata.SubType != "" {
		subType = event.Metadata.SubType
	}

	key := formatStorageKey(event.Metadata.Type, subType, event.Metadata.Version)
	vt, found := service.eventTypes[key]

	if found {
		return vt.Validate(event)
	}

	return fmt.Errorf("no matching schema found for: %s|%s|%s", event.Metadata.Type, event.Metadata.SubType, event.Metadata.Source)
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
					"eventType": eventType.Spec.EventType,
					"version":   eventType.Spec.Version,
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
					"eventType": eventType.Spec.EventType,
					"version":   eventType.Spec.Version,
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

	updated := false

	if len(eventType.Spec.SubTypes) > 0 {
		for _, subType := range eventType.Spec.SubTypes {
			if checkAndUpdate(eventType.Spec.EventType, subType, eventType.Spec.Version, service, eventType) {
				updated = true
			}
		}
		return updated
	} else {
		return checkAndUpdate(eventType.Spec.EventType, "*", eventType.Spec.Version, service, eventType)
	}
}

func onDeletedEventType(service *eventTypesExecutionService) func(eventTypeInterface interface{}) {
	return func(policyInterface interface{}) {
		eventType, successfulCast := policyInterface.(*types.EventType)
		if successfulCast {

			if len(eventType.Spec.SubTypes) > 0 {
				for _, key := range eventType.Spec.SubTypes {
					delete(service.eventTypes, formatStorageKey(eventType.Spec.EventType, key, eventType.Spec.Version))
				}
			} else {
				delete(service.eventTypes, formatStorageKey(eventType.Spec.EventType, "*", eventType.Spec.Version))
			}

			log.Logger.Info("deleted event type", zap.Any("eventType", map[string]string{
				"name":      eventType.Name,
				"namespace": eventType.Namespace,
				"eventType": eventType.Spec.EventType,
				"version":   eventType.Spec.Version,
			}))
		} else {
			log.Logger.Error("could not cast event type")
		}
	}
}

func formatStorageKey(eventType string, SubType string, eventVersion string) string {
	return fmt.Sprintf("%s|%s|%s", eventType, SubType, eventVersion)
}

func checkAndUpdate(eventType string, SubType string, eventVersion string, service *eventTypesExecutionService, eventTypeSchema *types.EventType) bool {

	key := formatStorageKey(eventType, SubType, eventVersion)

	schema, err := jsonSchema.CompileString("schema.json", eventTypeSchema.Spec.Schema)
	if err != nil {
		log.Logger.Error("Cannot parse json schema. Not adding schema to collection", zap.Any("eventType", map[string]string{
			"name":      eventTypeSchema.Name,
			"namespace": eventTypeSchema.Namespace,
			"eventType": eventTypeSchema.Spec.EventType,
			"version":   eventTypeSchema.Spec.Version,
		}), zap.Error(err))
		return false
	}

	et, found := service.eventTypes[key]
	if (found) && et.version == eventTypeSchema.ResourceVersion {
		return false
	}

	service.eventTypes[key] = validatableEventType{
		Schema:   *schema,
		SubTypes: eventTypeSchema.Spec.SubTypes,
		Sources:  eventTypeSchema.Spec.Sources,
		version:  eventTypeSchema.ResourceVersion,
	}

	return true
}
