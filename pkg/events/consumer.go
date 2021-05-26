package events

import (
	"github.com/gtforge/global_services_common_go/gett-mq"
	"github.com/gtforge/global_services_common_go/gett-mq/consumer"
	"github.com/gtforge/gorm"
	"github.com/sirupsen/logrus"
)

const (
	routingKey = "orders.update_rating"
	queueName = "orders.update_rating_queue"
)

type Events interface {
	ConsumeEvent() error
}

type Consume struct {
	RabbitMQ *gettMQ.AMQPConnection
}

func InitConsumer(db *gorm.DB) {
	ConsumeEvent(db)
}

func ConsumeEvent(db *gorm.DB) error{

	err := consumer.Subscribe(queueName, routingKey, GetConsumer(db))
	if err != nil {
		logrus.Error("error - unable to sent the event with personID: {} ")
		return err
	}

	return nil
}



