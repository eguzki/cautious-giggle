package utils

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func IgnoreAlreadyExists(err error) error {
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
