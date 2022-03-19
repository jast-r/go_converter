package goconverter

type Video struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	SourceFormat string `json:"src_format"`
	OutFormat    string `json:"out_format"`
}
