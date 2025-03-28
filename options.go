package webtest

type handle int

const (
	DEFAULT handle = iota
	SOFT
)

type WebClientOptions struct {
	WriteToFile bool   `json:"writeToFile"`
	FilePath    string `json:"filePath"`
	HandleError handle `json:"handleError"`
}
