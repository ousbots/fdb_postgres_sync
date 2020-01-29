package main

// Mutable represents the mutable appended data in fdb and postgres
type Mutable struct {
	ID        int64
	Data      []byte
	IDWritten bool
}

// containsByte returns true if the given Mutable struct contains the given byte, otherwise
// it returns false.
func (self Mutable) containsByte(element byte) bool {
	for index, _ := range self.Data {
		if self.Data[index] != element {
			return false
		}
	}

	return true
}
