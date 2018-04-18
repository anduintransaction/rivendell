package utils

// StringSet .
type StringSet map[string]struct{}

// NewStringSet .
func NewStringSet(values ...string) StringSet {
	s := make(StringSet)
	s.Add(values...)
	return s
}

// Add new values to string set
func (s StringSet) Add(values ...string) {
	for _, value := range values {
		s[value] = struct{}{}
	}
}

// Remove values from string set
func (s StringSet) Remove(values ...string) {
	for _, value := range values {
		delete(s, value)
	}
}

// Exists checks a value appears in the set or not
func (s StringSet) Exists(value string) bool {
	_, ok := s[value]
	return ok
}

// ToSlice converts the string set to a slice
func (s StringSet) ToSlice() []string {
	slice := []string{}
	for value := range s {
		slice = append(slice, value)
	}
	return slice
}

// First returns first element from the set
func (s StringSet) First() string {
	for value := range s {
		return value
	}
	return ""
}

// Join .
func (s StringSet) Join(another StringSet) StringSet {
	newStringSet := NewStringSet()
	for value := range s {
		if another.Exists(value) {
			newStringSet.Add(value)
		}
	}
	return newStringSet
}
