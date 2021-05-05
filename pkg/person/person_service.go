package person

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Service interface {
	GetPersons() ([]Person, error)
	CreatePersons(person *CreatePersonRequest) (*Person, error)
	GetPersonByID(id int64) (*Person, error)
	UpdatePerson(id int64, createPersonRequest *CreatePersonRequest) (*Person, error)
	DeletePerson(id int64) error
}

type PersonService struct {
	repository PersonRepository
	store      Provider
}

func NewPersonService(personRepository PersonRepository) Service {
	return &PersonService{
		repository: personRepository,
		store:      Instance,
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
		logrus.Error("error - persons", err)
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
		logrus.Error("error - person by id", err)
		return nil, err
	}
	logrus.Debug("persons num {}")

	return person, nil
}

func (s PersonService) GetPersons() ([]Person, error) {
	persons, err := s.store.GetPersons()
	if err == nil {
		return persons, nil
	}

	persons, err = s.repository.GetPersons()

	if err != nil {
		logrus.Error("error - persons", err)
		return nil, err
	}

	s.store.SetPersons(persons)
	logrus.Debug("persons num {}", len(persons))

	return persons, nil
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
