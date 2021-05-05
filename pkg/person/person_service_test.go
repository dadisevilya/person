package person

import (
	errors "github.com/ansel1/merry"
	"github.com/golang/mock/gomock"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"testing"
	"time"
)

func TestPersonService(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "person service test")
}

var _ = ginkgo.Describe("person service", func() {

	var (
		personService        *PersonService
		personsArray         []Person
		createPerson         CreatePersonRequest
		persons              []Person
		person               *Person
		err                  error
		ctrl                 *gomock.Controller
		personRepositoryMock *MockPersonRepository
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		personRepositoryMock = NewMockPersonRepository(ctrl)
		personService = &PersonService{
			repository: personRepositoryMock,
		}
		personsArray = make([]Person, 0)
	})

	var _ = ginkgo.Describe("getPerson Validations", func() {

		//before -> just before -> it
		ginkgo.JustBeforeEach(func() {
			persons, err = personService.GetPersons()
		})
		ginkgo.Context("validate that persons not return persons", func() {

			ginkgo.BeforeEach(func() {
				personRepositoryMock.EXPECT().GetPersons().Return(nil, errors.New("Invalid"))
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(persons).To(gomega.BeNil())
			})

		})
		ginkgo.Context("validate that persons return all persons", func() {

			ginkgo.BeforeEach(func() {
				personsArray = append(personsArray, Person{Name: "elad", Height: "179", Weight: "100", Age: 50})
				personsArray = append(personsArray, Person{Name: "lio", Height: "177", Weight: "110", Age: 28})
				personRepositoryMock.EXPECT().GetPersons().Return(personsArray, nil)
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(persons).To(gomega.Equal(personsArray))
			})
		})
	})

	var _ = ginkgo.Describe("getPersonByID Validations", func() {

		//before -> just before -> it
		ginkgo.JustBeforeEach(func() {
			person, err = personService.GetPersonByID(personsArray[0].ID)
		})

		ginkgo.BeforeEach(func() {
			personsArray = append(personsArray, Person{ID: 1, Name: "lio", Height: "177", Weight: "110", Age: 28})
			personsArray = append(personsArray, Person{ID: 2, Name: "lio", Height: "177", Weight: "110", Age: 28})
		})
		ginkgo.Context("validate that getPersonById not return person", func() {

			ginkgo.BeforeEach(func() {
				personRepositoryMock.EXPECT().GetPersonById(personsArray[0].ID).Return(nil, errors.New("Invalid"))
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(person).To(gomega.BeNil())
			})

		})
		ginkgo.Context("validate that persons return the specific person", func() {

			ginkgo.BeforeEach(func() {
				personRepositoryMock.EXPECT().GetPersonById(personsArray[0].ID).Return(&personsArray[0], nil)
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(person).To(gomega.Equal(&personsArray[0]))
			})
		})
	})

	var _ = ginkgo.Describe("deletePerson Validations", func() {

		//before -> just before -> it
		ginkgo.JustBeforeEach(func() {
			err = personService.DeletePerson(personsArray[0].ID)
		})

		ginkgo.BeforeEach(func() {
			personsArray = append(personsArray, Person{ID: 1, Name: "lio", Height: "177", Weight: "110", Age: 28})
			personsArray = append(personsArray, Person{ID: 2, Name: "lio", Height: "177", Weight: "110", Age: 28})
		})
		ginkgo.Context("validate that deletePerson not delete person", func() {

			ginkgo.BeforeEach(func() {
				personRepositoryMock.EXPECT().DeletePerson(personsArray[0].ID).Return(errors.New("can't delete"))
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.HaveOccurred())
			})

		})
		ginkgo.Context("validate that deletePerson delete person", func() {

			ginkgo.BeforeEach(func() {
				personRepositoryMock.EXPECT().DeletePerson(personsArray[0].ID).Return(nil)
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	var _ = ginkgo.Describe("createPerson Validations", func() {

		//before -> just before -> it
		ginkgo.JustBeforeEach(func() {
			person, err = personService.CreatePersons(&createPerson)
		})

		ginkgo.BeforeEach(func() {
			createPerson = CreatePersonRequest{Name: "dadi", Age: 26, Weight: "95", Height: "187"}
		})
		ginkgo.Context("validate that createPerson created a person", func() {

			ginkgo.BeforeEach(func() {

				personRepositoryMock.EXPECT().CreatePerson(gomock.Any()).Return(errors.New("can't create a person"))
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.HaveOccurred())
			})

		})
		ginkgo.Context("validate that createPerson not able to create person", func() {

			ginkgo.BeforeEach(func() {

				personRepositoryMock.EXPECT().CreatePerson(gomock.Any()).Return(nil)
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	var _ = ginkgo.Describe("updatePerson Validations", func() {

		var (
			per *Person
		)
		//before -> just before -> it
		ginkgo.JustBeforeEach(func() {
			person, err = personService.UpdatePerson(1, &createPerson)
		})

		ginkgo.BeforeEach(func() {
			createPerson = CreatePersonRequest{Name: "dadi", Age: 26, Weight: "95", Height: "187"}
		})
		ginkgo.Context("validate that updatePerson update the person", func() {

			ginkgo.BeforeEach(func() {
				per = &Person{1, createPerson.Name, createPerson.Age, createPerson.Height, createPerson.Weight, time.Now()}
				personRepositoryMock.EXPECT().GetPersonById(per.ID).Return(per, nil)
				personRepositoryMock.EXPECT().UpdatePerson(per.ID, &createPerson).Return(per, nil)
			})

			ginkgo.It("do", func() {
				gomega.Expect(person).To(gomega.Equal(per))
			})

		})
		ginkgo.Context("validate that updatePerson not update the person", func() {

			ginkgo.BeforeEach(func() {
				per = &Person{1, createPerson.Name, createPerson.Age, createPerson.Height, createPerson.Weight, time.Now()}
				personRepositoryMock.EXPECT().GetPersonById(per.ID).Return(per, nil)
				personRepositoryMock.EXPECT().UpdatePerson(per.ID, &createPerson).Return(nil, errors.New("unable to update person"))
			})

			ginkgo.It("do", func() {
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})

})
