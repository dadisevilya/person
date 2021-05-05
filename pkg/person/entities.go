package person

type CreatePersonRequest struct {
	Name   string `json:"name"`
	Age    int64  `json:"age"`
	Height string `json:"height"`
	Weight string `json:"weight"`
}
