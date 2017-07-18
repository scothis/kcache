package filter

import (
	"reflect"

	"github.com/boz/kcache/nsname"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/pkg/api/v1"
)

type Filter interface {

	// Accept() should return true if the given object passes the filter.
	Accept(metav1.Object) bool
}

type ComparableFilter interface {
	Filter
	Equals(Filter) bool
}

// Null() returns a filter whose Accept() is always true.
func Null() ComparableFilter {
	return nullFilter{}
}

type nullFilter struct{}

func (nullFilter) Accept(_ metav1.Object) bool {
	return true
}

func (nullFilter) Equals(other Filter) bool {
	_, ok := other.(nullFilter)
	return ok
}

type allFilter struct{}

// All() returns a filter whose Accept() is always false.
func All() ComparableFilter {
	return allFilter{}
}

func (allFilter) Accept(_ metav1.Object) bool {
	return false
}

func (allFilter) Equals(other Filter) bool {
	_, ok := other.(allFilter)
	return ok
}

// Labels() returns a filter which returns true if
// the provided map is a subset of the object's labels.
func Labels(match map[string]string) ComparableFilter {
	return &labelsFilter{match}
}

type labelsFilter struct {
	target map[string]string
}

func (f *labelsFilter) Accept(obj metav1.Object) bool {
	if len(f.target) == 0 {
		return true
	}

	current := obj.GetLabels()

	for k, v := range f.target {
		if val, ok := current[k]; !ok || val != v {
			return false
		}
	}
	return true
}

func (f *labelsFilter) Equals(other Filter) bool {
	if other, ok := other.(*labelsFilter); ok {
		if len(f.target) != len(other.target) {
			return false
		}
		if len(f.target) == 0 {
			return true
		}
		for k, v := range f.target {
			if val, ok := other.target[k]; !ok || val != v {
				return false
			}
		}
		return true
	}
	return false
}

// ServiceFor() returns a filter if the object
// is a Service whose selector mateches the given target.
func ServiceFor(target map[string]string) ComparableFilter {
	return &serviceForFilter{target}
}

type serviceForFilter struct {
	target map[string]string
}

func (f *serviceForFilter) Accept(obj metav1.Object) bool {
	svc, ok := obj.(*v1.Service)

	if !ok {
		return false
	}

	if len(svc.Spec.Selector) == 0 || len(f.target) == 0 {
		return false
	}

	for k, v := range svc.Spec.Selector {
		if val, ok := f.target[k]; !ok || val != v {
			return false
		}
	}

	return true
}

func (f *serviceForFilter) Equals(other Filter) bool {
	if other, ok := other.(*serviceForFilter); ok {
		return labels.Equals(f.target, other.target)
	}
	return false
}

// NSName() returns a filter whose Accept() returns true
// if the object's namespace and name matches one of the given
// NSNames.
func NSName(ids ...nsname.NSName) ComparableFilter {
	set := make(map[nsname.NSName]bool)
	for _, id := range ids {
		set[id] = true
	}
	return nsNameFilter(set)
}

type nsNameFilter map[nsname.NSName]bool

func (f nsNameFilter) Accept(obj metav1.Object) bool {
	key := nsname.ForObject(obj)
	_, ok := f[key]
	return ok
}

func (f nsNameFilter) Equals(other Filter) bool {
	return reflect.DeepEqual(f, other)
}
