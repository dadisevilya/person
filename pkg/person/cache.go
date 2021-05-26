package person

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type InMemoryProvider interface {
	GetInMemoryRatings() ([]Rating, error)
	SetInMemoryRatings(ratings []Rating) error
}

type Cache struct {
}

var CacheInstance = &Cache{}

func (c Cache) GetInMemoryRatings() (rating []Rating, err error) {
	file, _ := ioutil.ReadFile("ratings.json")

	err = json.Unmarshal([]byte(file), &rating)
	if err != nil {
		return rating, err
	}

	return rating,nil
}

func (c Cache) SetInMemoryRatings(ratings []Rating) error {
	file, _ := json.MarshalIndent(ratings, "", " ")
	err := ioutil.WriteFile("ratings.json", file, 0644)
	if err != nil {
		logrus.Error("unable write to file! ")
		return err
	}

	return nil
}
