package info

// NumDC indicates the number of datacenters
var NumDC int

// ID indicates the ID of this datacenter
var ID int

func InitChariots(numDc int, id int) {
	NumDC = numDc
	ID = id
}
