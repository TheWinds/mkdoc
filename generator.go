package docspace

type DocGenerator interface {
	// Set source object
	Source(api *API) DocGenerator
	// Generate doc
	Gen() (output string,err error)
}