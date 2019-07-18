package validator

// Validator is an interface for implementing a validator of a single
// Kubernetes object type. Ideally each Validator will check one aspect of
// an object, or perform several steps that have a common theme or goal.
type Validator interface {
	// Validate should run validation logic on an arbitrary object, and return
	// a one ManifestResult for each object that did not pass validation.
	// TODO: use pointers
	Validate() []ManifestResult
	// AddObjects adds objects to the Validator. Each object will be validated
	// when Validate() is called.
	AddObjects(...interface{}) Error
	// Name should return a succinct name for this validator.
	Name() string
}
