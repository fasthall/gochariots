package info

// NumDC indicates the number of datacenters
var NumDC int

// ID indicates the ID of this datacenter
var ID int

// InitChariots initializes the component.
// It should be called for every components in the cluster.
func InitChariots(numDc int, id int) {
	NumDC = numDc
	ID = id
}
