package person

import (
	"github.com/gtforge/gorm"
	"github.com/sirupsen/logrus"
	"strconv"
)

//mockgen -source=./pkg/person/person_repository.go -destination=./pkg/person/mock/person_repository_mock.go PersonRepository //-package=person_repository
type PersonRepository interface {
	GetPersons() ([]Person, error)
	CreatePerson(person *Person) error
	GetPersonById(id int64) (*Person, error)
	UpdatePerson(id int64, createPersonRequest *CreatePersonRequest) (*Person, error)
	UpdatePersonRating(id int64, updated bool) error
	DeletePerson(id int64) error
}

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) PersonRepository {
	return &Repo{
		db: db,
	}
}

func (r *Repo) CreatePerson(person *Person) error {
	return r.db.Create(person).Error
}

func (r *Repo) GetPersons() ([]Person, error) {
	persons := make([]Person, 0)

	if err := r.db.Find(&persons).Error; err != nil {
		logrus.Error("can't get persons ", err)
		return nil, err
	}

	return persons, nil
}

func (r *Repo) GetPersonById(id int64) (*Person, error) {
	person := Person{}

	if err := r.db.First(&person, id).Error; err != nil {
		logrus.Error("can't get persons", err)
		return nil, err
	}

	return &person, nil
}

func (r *Repo) UpdatePerson(id int64, createPersonRequest *CreatePersonRequest) (*Person, error) {
	p := Person{}

	if err := r.db.First(&p, id).Error; err != nil {
		logrus.Error("can't get persons", err)
		return nil, err
	}

	r.db.First(&p)
	p.Name = createPersonRequest.Name
	p.Age = createPersonRequest.Age
	p.Height = createPersonRequest.Height
	p.Weight = createPersonRequest.Weight
	p.RatingUpdated = createPersonRequest.RatingUpdate
	r.db.Save(&p)

	return &p, nil
}

func (r *Repo) UpdatePersonRating(id int64, updated bool) error {
	p := Person{}

	if err := r.db.Find(&p, "id = ?", strconv.FormatInt(id, 10)).Error; err != nil {
		logrus.Error("can't get person by id - update order ", err)
		return err
	}

	r.db.First(&p)

	p.RatingUpdated = updated
	r.db.Save(&p)

	return nil
}


func (r *Repo) DeletePerson(id int64) error {
	// using this to create object like in DB (with all the columns) and then pointer to this object
	// the query will look like -> DELETE FROM person WHERE id = {id};
	p := Person{}

	if err := r.db.First(&p, id).Error; err != nil {
		logrus.Error("can't delete person ", err)
		return err
	}

	return r.db.Delete(&p, id).Error
}
