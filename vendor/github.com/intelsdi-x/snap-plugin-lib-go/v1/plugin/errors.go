package plugin

import "fmt"

var (
	// ErrEmptyKey is returned when a Rule with an empty key is created
	ErrEmptyKey = fmt.Errorf("Key cannot be Empty")

	// ErrConfigNotFound is returned when a config doesn't exist in the config map
	ErrConfigNotFound = fmt.Errorf("config item not found")

	// ErrNotA<type> is returned when the found config item doesn't have the expected type
	ErrNotAString = fmt.Errorf("config item is not a string")
	ErrNotAnInt   = fmt.Errorf("config item is not an int64")
	ErrNotABool   = fmt.Errorf("config item is not a boolean")
	ErrNotAFloat  = fmt.Errorf("config item is not a float64")
)
