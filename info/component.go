package info

// Name is the name of this running component
var name string

func SetName(n string) {
	name = n
}

func GetName() string {
	return name
}

func WritePID() {
	// ioutil.WriteFile(name+".pid", []byte(strconv.Itoa(os.Getpid())+"\n"), 0644)
}
