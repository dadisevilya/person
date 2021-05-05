package person

import "time"

type Person struct {
	ID        int64     `json:"id, omitempty"`
	Name      string    `json:"name"`
	Age       int64     `json:"age"`
	Height    string    `json:"height"`
	Weight    string    `json:"weight"`
	CreatedAt time.Time `json:"created_at" sql:"type:time" gorm:"time"`
}

func (Person) TableName() string {
	return "persons"
}

//func (p *Person) GetValue() {
//
//}
