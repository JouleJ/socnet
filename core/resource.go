package core

import (
	"fmt"
	"regexp"
)

type Resource interface {
	Name() string
	Content() []byte
}

type ResourceManager interface {
	GetList() []Resource
}

func GetFirstResourceByRegexp(rm ResourceManager, reStr string) (Resource, error) {
	re, err := regexp.Compile(reStr)
	if err != nil {
		return nil, err
	}

	for _, r := range rm.GetList() {
		if re.Match([]byte(r.Name())) {
			return r, nil
		}
	}

	return nil, fmt.Errorf("Not found resource by regexp %v", reStr)
}
