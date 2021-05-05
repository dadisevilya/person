package skeleton

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// APIError implements MarshalJSON for better representing API errors
type APIError struct {
	message string
	error   error
}

// NewAPIError returns new APIError istance
func NewAPIError(message string, err error) APIError {
	return APIError{
		message: message,
		error:   err,
	}
}

func (e APIError) Error() string {
	return e.error.Error()
}

// SerializedAPIError structure to be serialized to represent APIError
type SerializedAPIError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// UnmarshalJSON custom serializer for APIError
func (e *APIError) UnmarshalJSON(b []byte) error {
	s := SerializedAPIError{}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal api error")
	}
	e.message = s.Message
	e.error = errors.New(s.Error)
	return nil
}

// MarshalJSON custom serializer for APIError
func (e APIError) MarshalJSON() ([]byte, error) {
	response := SerializedAPIError{}
	response.Message = e.message
	response.Error = e.error.Error()
	return json.Marshal(response)
}
