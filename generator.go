package docspace

// DocGenerator receive a API def and output a doc for this API def
type DocGenerator interface {
	// Set source object
	Source(api *API) DocGenerator
	// Generate doc
	Gen() (output string,err error)
}