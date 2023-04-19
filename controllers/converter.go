package controllers

import (
	"bytes"

	"github.com/manifestival/manifestival"
	"k8s.io/client-go/kubernetes/scheme"
)

func ConvertToStructuredResource(yamlContent []byte, out interface{}, opts ...manifestival.Option) error {
	reader := bytes.NewReader(yamlContent)
	m, err := manifestival.ManifestFrom(manifestival.Reader(reader), opts...)
	if err != nil {
		return err
	}

	s := scheme.Scheme
	RegisterSchemes(s)
	err = s.Convert(&m.Resources()[0], out, nil)
	if err != nil {
		return err
	}
	return nil
}
