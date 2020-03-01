package types

// List is the slice implemented Romove method
type List []interface{}

// Remove remove element at specified index.
func (l List) Remove(index int) List {
	if index == 0 {
		return l[1:]
	}
	var newList List
	newList = append(newList, l[:index]...)
	newList = append(newList, l[index+1:]...)
	return newList
}
