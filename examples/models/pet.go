package models

// Category groups pets by type (e.g., "Dogs", "Cats").
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Tag is a label attached to a pet (e.g., "friendly", "vaccinated").
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Pet represents a pet in the Petstore.
// Status can be "available", "pending", or "sold".
type Pet struct {
	ID        int64    `json:"id"`
	Category  Category `json:"category"`
	Name      string   `json:"name"`
	PhotoUrls []string `json:"photoUrls"`
	Tags      []Tag    `json:"tags"`
	Status    string   `json:"status"`
}
