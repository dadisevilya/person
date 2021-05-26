package person

import "time"

type Person struct {
	ID            int64     `json:"id, omitempty"`
	Name          string    `json:"name"`
	Age           int64     `json:"age"`
	Height        string    `json:"height"`
	Weight        string    `json:"weight"`
	RatingUpdated bool      `json:"rating_updated"`
	CreatedAt     time.Time `json:"created_at" sql:"type:time" gorm:"time"`
}

type Order struct {
	ID        int64     `json:"id, omitempty"`
	OrderID   int64     `json:"order_id"`
	PersonID  int64     `json:"person_id"`
	Rating    float64   `json:"rating"`
	CreatedAt time.Time `json:"created_at" sql:"type:time" gorm:"time"`
}

type Rating struct {
	PersonID int64   `json:"person_id"`
	Rating   float64 `json:"rating"`
}

func (Person) TableName() string {
	return "persons"
}

//func (p *Person) GetValue() {
//
//}
