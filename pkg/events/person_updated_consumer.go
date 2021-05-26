package events

import (
	"encoding/json"
	"github.com/gtforge/global_services_common_go/gett-mq"
	"github.com/gtforge/go-skeleton-draft/structure/pkg/person"
	"github.com/gtforge/gorm"
	"github.com/sirupsen/logrus"
)

//var PersonUpdatedConsumer = &personUpdatedConsumer{
//	service: person.NewPersonService(person.NewRepo(gettStorages.DB)),
//}

type personId struct {
	ID int64 `json:"id"`
}

func GetConsumer(db *gorm.DB) *personUpdatedConsumer{
	return &personUpdatedConsumer{
		service: person.NewPersonService(person.NewRepo(db)),
	}
}

type personUpdatedConsumer struct {
	service person.Service
}

func (p personUpdatedConsumer) Process(message gettMQ.MqMessage) error {
	event := personId{}
	err := json.Unmarshal(message.Payload, &event)
	if err != nil {
		logrus.Error("error - unable to unmarshal the update person event ", err)
	}


	err = p.service.UpdatePersonRating(event.ID, true)
	if err != nil {
		logrus.Error("error - unable to update person event ", err)
	}

	p.service.AsyncRun()

	return nil
}
