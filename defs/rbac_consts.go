package defs

type APIPermission struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}
