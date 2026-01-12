package flagconversion

import (
	"errors"

	"github.com/fastly/go-fastly/v12/fastly"
)

func ConvertBoolFromStringFlag(value string) (*bool, error) {
	switch value {
	case "true":
		return fastly.ToPointer(true), nil
	case "false":
		return fastly.ToPointer(false), nil
	default:
		return nil, errors.New("value must be one of the following [true, false]")
	}
}

func ConvertOrderFromStringFlag(value string) (string, error) {
	switch value {
	case "asc":
		return "", nil
	case "desc":
		return "-", nil
	default:
		return "", errors.New("value must be one of the following [asc, desc]")
	}
}
