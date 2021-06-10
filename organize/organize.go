package organize

import "fmt"

// Library object
type Library struct {
	InPath     string
	OutPath    string
	Topic      string
	DateFormat string
}

// Init variables
func (lib *Library) Init() {
	fmt.Println("Initializing variables")
}
