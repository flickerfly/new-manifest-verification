package validate

import (
	"github.com/dweepgogia/new-manifest-verification/pkg/validate/validator"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
)

type CRDValidator struct {
	crds []v1beta1.CustomResourceDefinition
}

var _ validator.Validator = &CRDValidator{}

func (v *CRDValidator) Validate() (results []validator.ManifestResult) {
	for _, crd := range v.crds {
		// Contains error logs for all missing optional and mandatory fields.
		//result := csvInspect(crd)
		result := crdInspect(crd)
		if result.Name == "" {
			result.Name = crd.GetName()
		}
		results = append(results, result)
	}
	return results
}

func (v *CRDValidator) AddObjects(objs ...interface{}) validator.Error {
	for _, o := range objs {
		switch t := o.(type) {
		case v1beta1.CustomResourceDefinition:
			v.crds = append(v.crds, t)
		case *v1beta1.CustomResourceDefinition:
			v.crds = append(v.crds, *t)
		}
	}
	return validator.Error{}
}

func (v CRDValidator) Name() string {
	return "ClusterServiceVersion Validator"
}

func crdInspect(crd v1beta1.CustomResourceDefinition) (manifestResult validator.ManifestResult) {
	scheme := runtime.NewScheme()
	err := apiextensions.AddToScheme(scheme)
	if err != nil {
		return
	}
	err = v1beta1.AddToScheme(scheme)
	if err != nil {
		return
	}
	unversionedCRD := &apiextensions.CustomResourceDefinition{}
	scheme.Converter().Convert(&crd, &unversionedCRD, conversion.SourceToDest, nil)
	errList := validation.ValidateCustomResourceDefinition(unversionedCRD)
	for _, err := range errList {
		er := validator.Error{Type: validator.ErrorType(err.Type), Field: err.Field, BadValue: err.BadValue, Detail: err.Detail}
		manifestResult.Errors = append(manifestResult.Errors, er)
	}
	return
}
