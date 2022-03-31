package flyteworkflow

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	preserveUnknownFields = true
	CRD                   = apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("flyteworkflows.%s", GroupName),
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: GroupName,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:       "FlyteWorkflow",
				Plural:     "flyteworkflows",
				Singular:   "flyteworkflow",
				ShortNames: []string{"fly"},
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				apiextensionsv1.CustomResourceDefinitionVersion{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: &preserveUnknownFields,
						},
					},
				},
			},
		},
	}
)
