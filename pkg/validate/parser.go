package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// "github.com/dweepgogia/new-manifest-verification/pkg/validate/validator"
	// "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	//"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry"
	// "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

type Manifest struct {
	Package registry.PackageManifest
	Bundle  []ManifestBundle
}

type ManifestBundle struct {
	Version string
	CRDs    []string
	CSV     string
}

func ParseDir(manifestDirectory string) Manifest {
	countPkg := 0
	manifest := Manifest{}
	// parsing directory structure
	_ = filepath.Walk(manifestDirectory, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.Contains(f.Name(), "package") {
			countPkg++
			if countPkg > 1 {
				return fmt.Errorf("More than one pkg in the manifest")
			}
		}

		if !f.IsDir() {
			if strings.Contains(f.Name(), "clusterserviceversion") {
				version := strings.Split(path, "/")
				manifest.Bundle = append(manifest.Bundle, ManifestBundle{Version: version[1], CSV: path})
			}

		}

		// if !strings.HasSuffix(path, ".yaml") {
		// 	return nil
		// }

		// err = validateResource(path, f, err)
		// if err != nil {
		// 	return err
		// }

		return nil
	})
	//fmt.Println("OUTSIDE WALK Count pkg ", countPkg, err)
	return manifest
}
