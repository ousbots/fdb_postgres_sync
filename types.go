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

// containsInt returns true if the given MutableData struct contains the given int64,
// otherwise false.
func (self MutableData) containsInt(element int64) bool {
	for _, existing := range self.Data {
		if existing == element {
			return true
		}
	}

	return false
}

func (self MutableData) deepCopy() *MutableData {
	var copied MutableData
	copied.ID = self.ID
	copied.Data = make([]int64, len(self.Data))
	copy(copied.Data, self.Data)
	copied.IDWritten = self.IDWritten

	return &copied
}

// deleteFirstElement delete the first occurence of the given element in self.Data
func (self *MutableData) deleteFirstElement(element int64) {
	var index int

	for index, _ = range self.Data {
		if self.Data[index] == element {
			self.Data = append(self.Data[:index], self.Data[index+1:]...)
			return
		}
	}
}
