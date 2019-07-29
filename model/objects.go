package model

type ObjectDirectory = map[string]DataObject

type DataObject struct {
	Path   string
	Data   interface{}
	Hash   string
	Cached string
}
