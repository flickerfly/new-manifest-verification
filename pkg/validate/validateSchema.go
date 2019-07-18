// The static validators defined in this file are taken as is
// from the operator-registry's codebase which can be found here
// https://github.com/operator-framework/operator-registry/blob/master/pkg/schema/schema.go
//
// Returned error types have been changed to generic error type defined by the
// Operator Verification Library.

package validate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	apiservervalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/dweepgogia/new-manifest-verification/pkg/validate/validator"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ValidateCRD(schemaFileName string, fileBytes []byte) *validator.ManifestResult {
	schemaBytes, err := ioutil.ReadFile(schemaFileName)
	if err != nil {
		return getManifestResult(validator.IOError(fmt.Sprintf("Error in reading %s file:  #%s ", schemaFileName, err), schemaFileName))
	}
	schemaBytesJson, err := yaml.YAMLToJSON(schemaBytes)
	if err != nil {
		return getManifestResult(validator.InvalidParse(fmt.Sprintf("Error parsing raw YAML to Json for %s file:  #%s ", schemaFileName, err), schemaFileName))
	}

	crd := v1beta1.CustomResourceDefinition{}
	json.Unmarshal(schemaBytesJson, &crd)

	exampleFileBytesJson, err := yaml.YAMLToJSON(fileBytes)
	if err != nil {
		return getManifestResult(validator.InvalidParse(fmt.Sprintf("Error parsing raw YAML to Json for fileBytes:  #%s ", err), ""))
	}
	unstructured := unstructured.Unstructured{}
	err = json.Unmarshal(exampleFileBytesJson, &unstructured)
	if err != nil {
		return getManifestResult(validator.InvalidParse(fmt.Sprintf("Error parsing unstructured Json bytes to %s:  #%s ", unstructured, err), ""))
	}

	// Validate CRD definition statically
	scheme := runtime.NewScheme()
	err = apiextensions.AddToScheme(scheme)
	if err != nil {
		return getManifestResult(validator.InvalidOperation(fmt.Sprintf("Error adding scheme:  #%s ", err))) // TODO: See if we need a specific function for these errors.
	}
	err = v1beta1.AddToScheme(scheme)
	if err != nil {
		return getManifestResult(validator.InvalidOperation(fmt.Sprintf("Error adding scheme:  #%s ", err))) // TODO: See if we need a specific function for these errors.
	}

	unversionedCRD := apiextensions.CustomResourceDefinition{}
	scheme.Converter().Convert(&crd, &unversionedCRD, conversion.SourceToDest, nil)
	errList := validation.ValidateCustomResourceDefinition(&unversionedCRD)
	if len(errList) > 0 {
		castErrList := []validator.Error{}
		for _, ferr := range errList {
			err := validator.Error{Type: validator.ErrorType(ferr.Type), Field: ferr.Field, BadValue: ferr.BadValue, Detail: ferr.Detail}
			castErrList = append(castErrList, err)
		}
		castErrList = append(castErrList, validator.FailedValidation(fmt.Sprintf("CRD failed validation: %s.", schemaFileName), schemaFileName))
		return &validator.ManifestResult{Errors: castErrList, Warnings: nil}
	}

	// Validate CR against CRD schema
	newSchemaValidator, _, err := apiservervalidation.NewSchemaValidator(unversionedCRD.Spec.Validation)
	err = apiservervalidation.ValidateCustomResource(unstructured.UnstructuredContent(), newSchemaValidator)
	if err != nil {
		return getManifestResult(validator.FailedValidation(err.Error(), schemaFileName))
	}
	return &validator.ManifestResult{}
}

func getManifestResult(errs ...validator.Error) *validator.ManifestResult {
	errList := append([]validator.Error{}, errs...)
	return &validator.ManifestResult{Errors: errList, Warnings: nil}
}

func ValidateKind(kind string, fileBytes []byte) *validator.ManifestResult {
	exampleFileBytesJson, err := yaml.YAMLToJSON(fileBytes)
	if err != nil {
		return getManifestResult(validator.InvalidParse(fmt.Sprintf("Error parsing raw YAML to Json for fileBytes:  #%s ", err), ""))
	}

	switch kind {
	case "ClusterServiceVersion":
		csv := v1alpha1.ClusterServiceVersion{}
		err = json.Unmarshal(exampleFileBytesJson, &csv)
		if err != nil {
			return getManifestResult(validator.InvalidParse(fmt.Sprintf("Error parsing unstructured Json bytes to %T:  #%s ", csv, err), ""))
		}
		err := validateExamplesAnnotations(&csv)
		if err != nil {
			return err
		}
		return nil
	case "CatalogSource":
		cs := v1alpha1.CatalogSource{}
		err = json.Unmarshal(exampleFileBytesJson, &cs)
		if err != nil {
			return getManifestResult(validator.InvalidParse(fmt.Sprintf("Error parsing unstructured Json bytes to %T:  #%s ", cs, err), ""))
		}
		return nil
	default:
		return getManifestResult(validator.InvalidOperation(fmt.Sprintf("Error didn't recognize validate-kind directive: %s", kind)))
	}
}

func validateExamplesAnnotations(csv *v1alpha1.ClusterServiceVersion) *validator.ManifestResult {
	var examples []v1beta1.CustomResourceDefinition
	var annotationsNames = []string{"alm-examples", "olm.examples"}
	var annotationsExamples string
	var ok bool
	annotations := csv.ObjectMeta.GetAnnotations()
	// Return right away if no examples annotations are found.
	if annotations == nil {
		return &validator.ManifestResult{}
	}
	// Expect either `alm-examples` or `old.examples` but not both
	// If both are present, `alm-examples` will be used
	for _, name := range annotationsNames {
		annotationsExamples, ok = annotations[name]
		if ok {
			break
		}
	}

	// Can't find examples annotations, simply return
	if annotationsExamples == "" {
		return &validator.ManifestResult{}
	}

	if err := json.Unmarshal([]byte(annotationsExamples), &examples); err != nil {
		return getManifestResult(validator.InvalidParse(fmt.Sprintf("Error parsing unstructured Json bytes to %T:  #%s ", examples, err), ""))
	}

	providedAPIs, err := getProvidedAPIs(csv)
	if err != nil {
		return err
	}
	parsedExamples, err := parseExamplesAnnotations(examples)
	if err != nil {
		return err
	}
	if matchGVKProvidedAPIs(parsedExamples, providedAPIs) != nil {
		return err
	}
	return nil
}

func getProvidedAPIs(csv *v1alpha1.ClusterServiceVersion) (map[schema.GroupVersionKind]struct{}, *validator.ManifestResult) {
	provided := map[schema.GroupVersionKind]struct{}{}

	for _, owned := range csv.Spec.CustomResourceDefinitions.Owned {
		parts := strings.SplitN(owned.Name, ".", 2)
		if len(parts) < 2 {
			return nil, getManifestResult(validator.InvalidParse(fmt.Sprintf("Error couldn't parse plural.group from crd name: %s", owned.Name), owned.Name))
		}
		provided[schema.GroupVersionKind{Group: parts[1], Version: owned.Version, Kind: owned.Kind}] = struct{}{}
	}

	for _, api := range csv.Spec.APIServiceDefinitions.Owned {
		provided[schema.GroupVersionKind{Group: api.Group, Version: api.Version, Kind: api.Kind}] = struct{}{}
	}
	return provided, nil
}

func parseExamplesAnnotations(examples []v1beta1.CustomResourceDefinition) (map[schema.GroupVersionKind]struct{}, *validator.ManifestResult) {
	parsed := map[schema.GroupVersionKind]struct{}{}
	for _, value := range examples {
		parts := strings.SplitN(value.APIVersion, "/", 2)
		if len(parts) < 2 {
			return nil, getManifestResult(validator.InvalidParse(fmt.Sprintf("Error couldn't parse group/version from crd kind: %s", value.Kind), value))
		}
		parsed[schema.GroupVersionKind{Group: parts[0], Version: parts[1], Kind: value.Kind}] = struct{}{}
	}
	return parsed, nil
}

func matchGVKProvidedAPIs(examples map[schema.GroupVersionKind]struct{}, providedAPIs map[schema.GroupVersionKind]struct{}) *validator.ManifestResult {
	for key := range examples {
		if _, ok := providedAPIs[key]; !ok {
			return getManifestResult(validator.FailedValidation(fmt.Sprintf("Error couldn't match %v in provided APIs list: %v", key, providedAPIs), providedAPIs))
		}
	}
	return nil
}
