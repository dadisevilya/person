package person

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Service interface {
	GetPersons() ([]Person, error)
	GetAllRatingsByWaitingGroups() ([]Rating, error)
	GetAllRatingsByChannels() ([]Rating, error)
	CreatePersons(person *CreatePersonRequest) (*Person, error)
	GetPersonByID(id int64) (*Person, error)
	GetRatingByPersonID(id int64) (float64, error)
	UpdatePerson(id int64, createPersonRequest *CreatePersonRequest) (*Person, error)
	UpdatePersonRating(id int64, updated bool) (error)
	DeletePerson(id int64) error
	AsyncRun()([]Rating, error)
}

type PersonService struct {
	repository PersonRepository
	store      Provider
	cache 	   InMemoryProvider
}

func NewPersonService(personRepository PersonRepository) Service {
	return &PersonService{
		repository: personRepository,
		store:      Instance,
		cache:      CacheInstance,
	}
}

func (s PersonService) CreatePersons(person *CreatePersonRequest) (*Person, error) {
	p := Person{
		Name:      person.Name,
		Age:       person.Age,
		Weight:    person.Weight,
		Height:    person.Height,
		CreatedAt: time.Now(),
	}

	if err := s.repository.CreatePerson(&p); err != nil {
		return nil, err
	}
	s.store.CreatePersons(&p)

	logrus.Debug("get the new persons value")
	persons, err := s.repository.GetPersons()

	if err != nil {
		logrus.Error("error - persons ", err)
		return nil, err
	}

	s.store.SetPersons(persons)
	return &p, nil
}

func (s PersonService) GetPersonByID(id int64) (*Person, error) {
	person, err := s.store.GetPersonByID(id)
	if err == nil {
		return person, nil
	}

	person, err = s.repository.GetPersonById(id)

	if err != nil {
		logrus.Error("error - person by id ", err)
		return nil, err
	}
	logrus.Debug("persons num {}")

	return person, nil
}

func (s PersonService) GetRatingByPersonID(id int64) (float64, error) {
	orders := make([]Order, 0)
	response, err := http.Get(fmt.Sprintf("http://localhost:8081/api/v1/order_by_person/%v", id))
	if err != nil {
		return 0, err
	}

	body, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		panic(err.Error())
	}

	err = json.Unmarshal(body, &orders)
	if err != nil {
		logrus.Error("unable unmarshal get persons", err)
		return 0, err
	}

	if len(orders) == 0 {
		return 0.0, nil
	}

	var sum float64
	for _, u := range orders {
		sum += u.Rating
	}

	return sum / float64(len(orders)), nil
}

func (s PersonService) GetPersons() ([]Person, error) {
	persons, err := s.store.GetPersons()
	if err == nil {
		return persons, nil
	}

	persons, err = s.repository.GetPersons()

	if err != nil {
		logrus.Error("error - persons ", err)
		return nil, err
	}

	s.store.SetPersons(persons)
	logrus.Debug("persons num {}", len(persons))

	return persons, nil
}

func (s PersonService) GetAllRatingsByWaitingGroups() ([]Rating, error) {
	inMemoryRatings, _ := s.cache.GetInMemoryRatings()
	if len(inMemoryRatings) > 0 {
		return inMemoryRatings, nil
	}

	ratings, _ := s.AsyncRun()

	s.cache.SetInMemoryRatings(ratings)

	return ratings, nil
}

func (s PersonService) GetAllRatingsByChannels() ([]Rating, error) {
	persons, err := s.GetPersons()
	if err != nil {
		logrus.Error("error - persons ", err)
		return []Rating{}, err
	}

	c := make(chan Rating)
	for _, person := range persons {
		go s.GetRatingChannelByPersonID(person, c)

	}
	result := make([]Rating, len(persons))
	for i, _ := range result {
		result[i] = <-c
		fmt.Println(fmt.Sprintf("personID: %v rating: %v", result[i].PersonID, result[i].Rating))
	}

	return result, nil
}

func (s PersonService) GetRatingChannelByPersonID(person Person, c chan Rating) {
	orders := make([]Order, 0)
	response, err := http.Get(fmt.Sprintf("http://localhost:8081/api/v1/order_by_person/%v", person.ID))
	if err != nil {
		c <- Rating{person.ID, 0.0}
	}

	body, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		c <- Rating{person.ID, 0.0}
	}

	err = json.Unmarshal(body, &orders)
	if err != nil {
		c <- Rating{person.ID, 0.0}
	}

	if len(orders) == 0 {
		c <- Rating{person.ID, 0.0}
	}

	var sum float64
	for _, u := range orders {
		sum += u.Rating
	}

	c <- Rating{person.ID, sum / float64(len(orders))}
}

func (s PersonService) UpdatePerson(id int64, createPersonRequest *CreatePersonRequest) (*Person, error) {
	person, err := s.repository.GetPersonById(id)
	if err != nil {
		logrus.Error("error - person by id", err)
		return nil, err
	}

	personUpdated, err := s.repository.UpdatePerson(person.ID, createPersonRequest)
	if err != nil {
		logrus.Error("error - unable to update person with personID: {}", person.ID)
		return nil, err
	}

	_, err = s.store.UpdatePerson(person.ID, createPersonRequest)
	if err != nil {
		logrus.Error("error - update person from redis fails!!!", err)
		return nil, err
	}

	persons, err := s.repository.GetPersons()

	if err != nil {
		logrus.Error("error - persons", err)
		return nil, err
	}

	s.store.SetPersons(persons)
	return personUpdated, nil
}

func (s PersonService) UpdatePersonRating(id int64, updated bool) error {
	err := s.repository.UpdatePersonRating(id, updated)
	if err != nil {
		logrus.Error("error - unable to update person with personID: {}", id)
		return err
	}

	return nil
}


func (s PersonService) DeletePerson(id int64) error {
	err := s.store.DeletePerson(id)
	if err != nil {
		return err
	}

	err = s.repository.DeletePerson(id)

	if err != nil {
		logrus.Error("error - delete person by id", err)
		return err
	}

	logrus.Debug("get the new persons value")
	persons, err := s.repository.GetPersons()

	if err != nil {
		logrus.Error("error - persons", err)
		return err
	}

	s.store.SetPersons(persons)
	return nil
}

func (s PersonService) AsyncRun()([]Rating, error)  {
	persons, err := s.GetPersons()
	if err != nil {
		logrus.Error("error - persons ", err)
		return []Rating{}, err
	}

	ratings := make([]Rating, 0)
	var wg sync.WaitGroup
	for _, u := range persons {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			r, _ := s.GetRatingByPersonID(id)
			ratings = append(ratings, Rating{id, r})
		}(u.ID)
	}
	wg.Wait()

	return ratings, nil
}
