package person

import (
	"encoding/json"
	"github.com/gtforge/global_services_common_go/gett-storages"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type Provider interface {
	GetPersons() ([]Person, error)
	CreatePersons(person *Person)
	SetPersons([]Person)
	GetPersonByID(id int64) (*Person, error)
	UpdatePerson(id int64, createPersonRequest *CreatePersonRequest) (*Person, error)
	DeletePerson(id int64) error
}

var Instance = &PersonStore{}

type PersonStore struct {
}

func (ps PersonStore) GetPersons() ([]Person, error) {
	bytes, err := gettStorages.RedisClient.Get("persons").Bytes()

	if err != nil {
		logrus.Error("couldn't get redis get persons ", err)
		return []Person{}, err
	}
	persons := []Person{}
	err = json.Unmarshal(bytes, &persons)
	if err != nil {
		logrus.Error("unable unmarshal get persons", err)
		return []Person{}, err
	}

	return persons, nil
}

func (ps PersonStore) GetPersonByID(personId int64) (*Person, error) {
	bytes, err := gettStorages.RedisClient.Get("person:" + strconv.FormatInt(personId, 10)).Bytes()

	if err != nil {
		logrus.Error("couldn't get redis get person by ID ", err)
		return &Person{}, err
	}
	person := Person{}
	err = json.Unmarshal(bytes, &person)
	if err != nil {
		logrus.Error("unable unmarshal get person by ID", err)
		return &Person{}, err
	}

	return &person, nil
}

func (ps PersonStore) SetPersons(persons []Person) {
	bytes, err := json.Marshal(persons)
	if err != nil {
		logrus.Error("unable marshal get persons", err)
	}
	pipe := gettStorages.RedisClient.Pipeline()
	pipe.Set("persons", bytes, 25*time.Hour)

	_, err = pipe.Exec()
	if err != nil {
		logrus.Error("redis sucks!!", err)
	}
}

func (ps PersonStore) CreatePersons(person *Person) {
	bytes, err := json.Marshal(person)
	if err != nil {
		logrus.Error("unable marshal get persons", err)
	}
	pipe := gettStorages.RedisClient.Pipeline()
	pipe.Set("person:"+strconv.FormatInt(person.ID, 10), bytes, 25*time.Hour)

	_, err = pipe.Exec()
	if err != nil {
		logrus.Error("redis sucks!!", err)
	}
}

func (ps PersonStore) UpdatePerson(id int64, createPersonRequest *CreatePersonRequest) (*Person, error) {
	person, err := ps.GetPersonByID(id)
	if err != nil {
		logrus.Error("unable marshal get persons ", err)
		return &Person{}, err
	}

	person.Name = createPersonRequest.Name
	person.Age = createPersonRequest.Age
	person.Weight = createPersonRequest.Weight
	person.Height = createPersonRequest.Height

	bytes, err := json.Marshal(person)
	if err != nil {
		logrus.Error("unable marshal get persons", err)
	}

	pipe := gettStorages.RedisClient.Pipeline()
	pipe.Set("person:"+strconv.FormatInt(id, 10), bytes, 25*time.Hour)

	_, err = pipe.Exec()
	if err != nil {
		logrus.Error("redis sucks!!", err)
	}
	return person, nil
}

func (ps PersonStore) DeletePerson(id int64) error {
	if err := gettStorages.RedisClient.Del("person:" + strconv.FormatInt(id, 10)); err != nil {
		logrus.Error("can't delete person from redis ", err)
		return err.Err()
	}

	//remove persons
	if err := gettStorages.RedisClient.Del("persons"); err != nil {
		logrus.Error("can't delete persons from redis ", err)
		return err.Err()
	}
	return ps.DeletePerson(id)
}
