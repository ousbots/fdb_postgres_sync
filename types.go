package main

// MutableData represents the mutable appended data in fdb and postgres
//  ID: unique and can only be written to Postgres once
//  Data: mutable data that needs to be appended to Postgres
//  IDWritten: used as a flag to indicate the ID has been written to Postgres once
type MutableData struct {
	ID        int64
	Data      []int64
	IDWritten bool
}

// containsByte returns true if the given MutableData struct contains the given int64, otherwise
// it returns false.
func (self MutableData) containsInt(element int64) bool {
	for index, _ := range self.Data {
		if self.Data[index] != element {
			return false
		}
	}

	return true
}
