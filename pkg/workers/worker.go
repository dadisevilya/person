package workers

import (
	"github.com/gtforge/global_services_common_go/gett-storages"
	"github.com/gtforge/global_services_common_go/gett-workers"
	"github.com/gtforge/go-skeleton-draft/structure/pkg/person"
	"github.com/sirupsen/logrus"
)
import "github.com/gtforge/go-workers"

var CacheWorker = InMemoryWorker{}

type InMemoryWorker struct {
	gettWorkers.BaseWorker
	personService person.Service
	cache         person.Cache
}

func GetWorker() gettWorkers.Worker {
	return &InMemoryWorker{
		BaseWorker:    gettWorkers.BaseWorker{},
		personService: person.NewPersonService(person.NewRepo(gettStorages.DB)),
		cache:         *person.CacheInstance,
	}
}

func (imw InMemoryWorker) GetWorkerOptions() gettWorkers.WorkerOptions {
	return gettWorkers.WorkerOptions{
		Name:        "InMemoryWorker",
		Params:      map[string]interface{}{},
		CronLine:    "*/1 * * * *",
		Unique:      true,
		UniqueKey:   "InMemoryWorker",
		Concurrency: 200,
	}
}

func (imw InMemoryWorker) Perform(params *workers.Msg) {
	logrus.Info("start to perform... ")
	ratings, err := imw.personService.AsyncRun()
	if err != nil {
		logrus.Error("fail in perform ")
		return
	}
	err = imw.cache.SetInMemoryRatings(ratings)
	if err != nil {
		logrus.Error("unable to set the file ")
		return
	}
}
