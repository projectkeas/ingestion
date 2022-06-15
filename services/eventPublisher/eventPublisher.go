package eventPublisher

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/projectkeas/sdks-service/configuration"
	log "github.com/projectkeas/sdks-service/logger"
	"go.uber.org/zap"

	cejsm "github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	SERVICE_NAME string = "EventPublisher"
)

type EventPublisherService interface {
	Publish(event cloudevents.Event) bool
}

type eventPublisherExecutionService struct {
	natsClientCache map[string]cloudevents.Client
	config          *configuration.ConfigurationRoot
	mutex           *sync.Mutex
}

func New(config *configuration.ConfigurationRoot) EventPublisherService {
	service := eventPublisherExecutionService{
		mutex: &sync.Mutex{},
	}
	config.RegisterChangeNotificationHandler(func(c configuration.ConfigurationRoot) {

		service.mutex.Lock()
		defer service.mutex.Unlock()

		service.config = &c
		service.natsClientCache = map[string]cloudevents.Client{}
	})

	return &service
}

func (ep *eventPublisherExecutionService) Publish(event cloudevents.Event) bool {
	err := event.Validate()
	if err != nil {
		log.Logger.Error("Unable to validate outbound CloudEvent", zap.Error(err))
		return false
	}

	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	if ep.natsClientCache == nil {
		ep.natsClientCache = map[string]cloudevents.Client{}
	}

	streamName, subject := getStreamConfig(event.Type())
	client, found := ep.natsClientCache[subject]
	if !found {
		address := ep.config.GetStringValueOrDefault("nats.address", "nats-cluster.svc.cluster.local")
		port := ep.config.GetStringValueOrDefault("nats.port", "4222")
		uri := fmt.Sprintf("%s:%s", address, port)
		sender, err := cejsm.NewSender(uri, streamName, subject, cejsm.NatsOptions(), nil)
		if err != nil {
			log.Logger.Error("Unable to create new JetStream sender", zap.Error(err))
			return false
		}

		subjectWildcard := getSubjectWildcard(subject)
		streamInfo, err := sender.Jsm.StreamInfo(streamName)
		if err != nil {
			log.Logger.Error("Unable to update stream information", zap.Error(err))
			return false
		}

		if streamInfo != nil && !contains(streamInfo.Config.Subjects, subjectWildcard) {
			// TODO :: Add duplication window to config
			// TODO :: Add max age to config
			sender.Jsm.UpdateStream(&nats.StreamConfig{
				Name:        streamName,
				Subjects:    append(streamInfo.Config.Subjects, subjectWildcard),
				Description: streamInfo.Config.Description,
				Duplicates:  5 * time.Minute,
				Retention:   streamInfo.Config.Retention,
				MaxAge:      7 * (24 * time.Hour),
			})
		}

		client, err = cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		if err != nil {
			log.Logger.Error("Unable to create new CloudEvents Client", zap.Error(err))
			return false
		}

		ep.natsClientCache[subject] = client
	}

	result := client.Send(context.Background(), event)
	if cloudevents.IsUndelivered(result) {
		log.Logger.Error("Unable to publish to JetStream Cluster", zap.Error(result), zap.Any("nats", map[string]interface{}{
			"stream":       streamName,
			"subject":      subject,
			"uuid":         event.ID(),
			"acknowledged": cloudevents.IsACK(result),
		}))
	} else {
		log.Logger.Debug("Sent event", zap.Any("nats", map[string]interface{}{
			"stream":       streamName,
			"subject":      subject,
			"uuid":         event.ID(),
			"acknowledged": cloudevents.IsACK(result),
		}))
	}

	return result == nil
}

func getStreamConfig(input string) (string, string) {
	result := strings.Split(input, ".")
	length := len(result)

	if length > 1 {
		return result[1], strings.Join(result[1:], ".")
	}

	return input, input
}

func getSubjectWildcard(input string) string {
	sections := strings.Split(input, ".")
	joined := strings.Join(sections[0:len(sections)-1], ".")

	return joined + ".*"
}

func contains(elements []string, item string) bool {
	for _, element := range elements {
		if element == item {
			return true
		}
	}
	return false
}
