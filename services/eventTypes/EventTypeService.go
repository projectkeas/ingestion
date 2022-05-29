package eventTypes

import (
	"fmt"
	"time"

	types "github.com/projectkeas/crds/pkg/apis/keas.io/v1alpha1"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/ingestion/services"
	log "github.com/projectkeas/sdks-service/logger"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

const (
	SERVICE_NAME string = "EventTypes"
)

type EventTypeService interface {
	Validate(event sdk.EventEnvelope) bool
}

type eventTypesExecutionService struct {
	eventTypes map[string]validatableEventType
}

func (service eventTypesExecutionService) Validate(event sdk.EventEnvelope) bool {

	key := formatStorageKey(event.Metadata.EventType, event.Metadata.EventSubType, event.Metadata.EventVersion)
	vt, found := service.eventTypes[key]

	if found {
		return vt.Validate(event)
	}

	return false
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
			addOrUpdateEventType(service, eventType)
			log.Logger.Info("added new event type", zap.Any("eventType", map[string]string{
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

func onUpdatedEventType(service *eventTypesExecutionService) func(oldEventTypeInterface interface{}, newEventTypeInterface interface{}) {
	return func(oldEventType interface{}, newEventType interface{}) {
		eventType, successfulCast := newEventType.(*types.EventType)
		if successfulCast {
			addOrUpdateEventType(service, eventType)
			log.Logger.Info("updated event type", zap.Any("eventType", map[string]string{
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

func addOrUpdateEventType(service *eventTypesExecutionService, eventType *types.EventType) {

	if service.eventTypes == nil {
		service.eventTypes = map[string]validatableEventType{}
	}

	if len(eventType.Spec.SubTypes) > 0 {
		for _, subType := range eventType.Spec.SubTypes {
			key := formatStorageKey(eventType.Spec.EventType, subType, eventType.Spec.Version)
			service.eventTypes[key] = validatableEventType{
				Schema:   eventType.Spec.Schema,
				SubTypes: eventType.Spec.SubTypes,
				Sources:  eventType.Spec.Sources,
			}
		}
	} else {
		key := formatStorageKey(eventType.Spec.EventType, "*", eventType.Spec.Version)
		service.eventTypes[key] = validatableEventType{
			Schema:   eventType.Spec.Schema,
			SubTypes: eventType.Spec.SubTypes,
			Sources:  eventType.Spec.Sources,
		}
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

func formatStorageKey(eventType string, eventSubType string, eventVersion string) string {
	return fmt.Sprintf("%s|%s|%s", eventType, eventSubType, eventVersion)
}
