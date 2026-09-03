package models

// Order represents a purchase order for a pet.
// Status can be "placed", "approved", or "delivered".
type Order struct {
	ID       int64  `json:"id"`
	PetID    int64  `json:"petId"`
	Quantity int    `json:"quantity"`
	Status   string `json:"status"`
	Complete bool   `json:"complete"`
}
