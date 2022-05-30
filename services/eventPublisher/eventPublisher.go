package eventPublisher

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/sdks-service/configuration"
	log "github.com/projectkeas/sdks-service/logger"
	"go.uber.org/zap"
)

const (
	SERVICE_NAME string = "EventPublisher"
)

type EventPublisherService interface {
	Publish(metadata sdk.EventMetadata, data []byte) bool
}

type eventPublisherExecutionService struct {
	natsClient *nats.Conn
}

func New(config *configuration.ConfigurationRoot) EventPublisherService {
	service := eventPublisherExecutionService{}
	config.RegisterChangeNotificationHandler(func(c configuration.ConfigurationRoot) {
		existingClient := service.natsClient
		address := c.GetStringValueOrDefault("nats.address", "nats-cluster.svc.cluster.local")
		port := c.GetStringValueOrDefault("nats.port", "4222")
		uri := fmt.Sprintf("%s:%s", address, port)

		token := c.GetStringValueOrDefault("nats.token", "")

		nc, err := nats.Connect(uri, nats.Name("ingestion"), nats.Token(token))
		if err == nil {
			service.natsClient = nc
			if existingClient != nil {
				if log.Logger != nil {
					log.Logger.Info("Connection to the NATS cluster has been refreshed as the connection details have changed")
				}
				existingClient.Close()
			}
		} else {
			if log.Logger != nil {
				log.Logger.Error("Unable to establish connection to NATS cluster", zap.Error(err))
			}
		}
	})

	return &service
}

func (ep *eventPublisherExecutionService) Publish(metadata sdk.EventMetadata, data []byte) bool {

	if ep.natsClient == nil {
		if log.Logger != nil {
			log.Logger.Error("No connection to NATS cluster available")
		}
		return false
	}

	err := ep.natsClient.Publish(fmt.Sprintf("%s.%s", metadata.Source, metadata.Type), data)

	if err != nil {
		if log.Logger != nil {
			log.Logger.Error("Unable to publish to NATS cluster", zap.Error(err))
		}
		return false
	}

	return true
}

func (ep *eventPublisherExecutionService) Dispose() {
	if ep.natsClient != nil {
		ep.natsClient.Close()
		ep.natsClient = nil
	}
}
