package model

type ObjectDirectory = map[string]DataObject

type DataObject struct {
	Path   string
	Data   interface{}
	Error  error
	Hash   string
	Cached string
}
