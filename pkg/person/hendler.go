package person

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
	"net/http"
	"strconv"
)

type handler struct {
	service Service
	render  *render.Render
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
		render:  render.New(),
	}
}

func (h handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/persons", h.GetPersons).Methods(http.MethodGet)
	router.HandleFunc("/person/{id:[0-9]+}", h.GetPersonByID).Methods(http.MethodGet)
	router.HandleFunc("/rating/{id:[0-9]+}", h.GetRatingByPersonID).Methods(http.MethodGet)
	router.HandleFunc("/ratings/groups", h.GetAllRatingsByWaitingGroups).Methods(http.MethodGet)
	router.HandleFunc("/ratings/channels", h.GetAllRatingsByChannels).Methods(http.MethodGet)
	router.HandleFunc("/person", h.CreatePerson).Methods(http.MethodPost)
	router.HandleFunc("/update_person/{id:[0-9]+}", h.UpdatePerson).Methods(http.MethodPut)
	router.HandleFunc("/delete_person/{id:[0-9]+}", h.DeletePerson).Methods(http.MethodDelete)
}

func (h handler) CreatePerson(w http.ResponseWriter, req *http.Request) {
	logrus.WithFields(logrus.Fields{"vars": mux.Vars(req)}).Debug("create person")
	defer req.Body.Close()
	createPersonRequest := &CreatePersonRequest{}
	decoder := json.NewDecoder(req.Body)

	if err := decoder.Decode(&createPersonRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	person, err := h.service.CreatePersons(createPersonRequest)

	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	h.render.JSON(w, http.StatusCreated, person)
}

func (h handler) GetPersons(w http.ResponseWriter, req *http.Request) {
	logrus.WithFields(logrus.Fields{"vars": mux.Vars(req)}).Debug("get persons")
	persons, err := h.service.GetPersons()

	if err != nil {
		h.render.JSON(w, http.StatusNotFound, err)
		return
	}
	h.render.JSON(w, http.StatusOK, persons)
}

func (h handler) GetAllRatingsByWaitingGroups(w http.ResponseWriter, req *http.Request) {
	logrus.WithFields(logrus.Fields{"vars": mux.Vars(req)}).Debug("get all ratings")
	persons, err := h.service.GetAllRatingsByWaitingGroups()

	if err != nil {
		h.render.JSON(w, http.StatusNotFound, err)
		return
	}
	h.render.JSON(w, http.StatusOK, persons)
}

func (h handler) GetAllRatingsByChannels(w http.ResponseWriter, req *http.Request) {
	logrus.WithFields(logrus.Fields{"vars": mux.Vars(req)}).Debug("get all ratings")
	persons, err := h.service.GetAllRatingsByChannels()

	if err != nil {
		h.render.JSON(w, http.StatusNotFound, err)
		return
	}
	h.render.JSON(w, http.StatusOK, persons)
}

func (h handler) GetPersonByID(w http.ResponseWriter, req *http.Request) {
	logrus.WithFields(logrus.Fields{"vars": mux.Vars(req)}).Debug("get person by id")
	params := mux.Vars(req)
	stringId := params["id"]
	id, _ := strconv.ParseInt(stringId, 10, 64)

	person, err := h.service.GetPersonByID(id)

	if err != nil {
		h.render.JSON(w, http.StatusNotFound, err)
		return
	}
	h.render.JSON(w, http.StatusOK, person)
}

func (h handler) GetRatingByPersonID(w http.ResponseWriter, req *http.Request) {
	logrus.WithFields(logrus.Fields{"vars": mux.Vars(req)}).Debug("get rating by person id")
	params := mux.Vars(req)
	stringId := params["id"]
	id, _ := strconv.ParseInt(stringId, 10, 64)

	person, err := h.service.GetRatingByPersonID(id)

	if err != nil {
		h.render.JSON(w, http.StatusNotFound, err)
		return
	}
	h.render.JSON(w, http.StatusOK, person)
}

func (h handler) UpdatePerson(w http.ResponseWriter, req *http.Request) {
	logrus.WithFields(logrus.Fields{"vars": mux.Vars(req)}).Debug("update person by id")
	defer req.Body.Close()

	createPersonRequest := &CreatePersonRequest{}
	decoder := json.NewDecoder(req.Body)

	if err := decoder.Decode(&createPersonRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params := mux.Vars(req)
	stringId := params["id"]
	id, _ := strconv.ParseInt(stringId, 10, 64)

	person, err := h.service.UpdatePerson(id, createPersonRequest)

	if err != nil {
		h.render.Text(w, http.StatusNotFound, err.Error())
		return
	}
	h.render.JSON(w, http.StatusCreated, person)
}

func (h handler) DeletePerson(w http.ResponseWriter, req *http.Request) {
	logrus.WithFields(logrus.Fields{"vars": mux.Vars(req)}).Debug("delete person by id")
	defer req.Body.Close()

	params := mux.Vars(req)
	stringId := params["id"]
	id, _ := strconv.ParseInt(stringId, 10, 64)

	err := h.service.DeletePerson(id)

	if err != nil {
		h.render.Text(w, http.StatusNotFound, err.Error())
		return
	}
	h.render.JSON(w, http.StatusOK, "person with id: "+stringId)

}
