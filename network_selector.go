package main

import "fmt"

type selector struct {
	includes map[string]struct{}
	excludes map[string]struct{}
}

func NewDeviceSelector(includes, excludes []string) (*selector, error) {
	if len(includes) > 0 {
		if (len(excludes) == 1 && excludes[0] != getLocalInterfaceName()) || len(excludes) > 1 {
			return nil, fmt.Errorf("includes and excludes can't have interfaces simultaneously")
		}
	}
	selector := &selector{
		includes: map[string]struct{}{},
		excludes: map[string]struct{}{},
	}
	for _, include := range includes {
		selector.includes[include] = struct{}{}
	}
	for _, exclude := range excludes {
		selector.excludes[exclude] = struct{}{}
	}

	return selector, nil
}

// Ignored returns whether the device should be Ignored
func (ds *selector) Ignored(name string) bool {
	if len(ds.includes) > 0 {
		_, ok := ds.includes[name]
		return !ok
	}
	if len(ds.excludes) > 0 {
		_, ok := ds.excludes[name]
		return ok
	}
	return false
}
