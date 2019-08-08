package model

type ObjectDirectory = map[string]DataObject

type DataObject struct {
	Path   string      `json:"path"`
	Data   interface{} `json:"data"`
	Error  error       `json:"error"`
	Hash   string      `json:"hash"`
	Cached string      `json:"cached"`
}
