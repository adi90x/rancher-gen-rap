package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"
	"time"
	"reflect"

	log "github.com/Sirupsen/logrus"
)
//RAP : GroupbyMulti ( from jwilder/dockergen)
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
//RAP : dict ( from jwilder/dockergen)
func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}
//RAP : arrayClosest ( from jwilder/dockergen)
func arrayClosest(values []string, input string) string {
	best := ""
	for _, v := range values {
		if strings.Contains(input, v) && len(v) > len(best) {
			best = v
		}
	}
	return best
}
//RAP : arrayFirst ( from jwilder/dockergen)
func arrayFirst(input interface{}) interface{} {
	if input == nil {
		return nil
	}

	arr := reflect.ValueOf(input)

	if arr.Len() == 0 {
		return nil
	}

	return arr.Index(0).Interface()
}
//RAP : coalesce ( from jwilder/dockergen)
func coalesce(input ...interface{}) interface{} {
	for _, v := range input {
		if v != nil {
			return v
		}
	}
	return nil
}
func newFuncMap(ctx *TemplateContext) template.FuncMap {
	return template.FuncMap{
		// Utility funcs
		"base":      path.Base,
		"dir":       path.Dir,
		"env":       os.Getenv,
		"timestamp": time.Now,
		"split":     strings.Split,
		"join":      strings.Join,
		"toUpper":   strings.ToUpper,
		"toLower":   strings.ToLower,
		"contains":  strings.Contains,
		"replace":   strings.Replace,

		// Service funcs
		"host":              hostFunc(ctx),
		"hosts":             hostsFunc(ctx),
		"service":           serviceFunc(ctx),
		"services":          servicesFunc(ctx),
		"whereLabelExists":  whereLabelExists,
		"whereLabelEquals":  whereLabelEquals,
		"whereLabelMatches": whereLabelEquals,
		"groupByLabel":      groupByLabel,
		
		//Add for Rancher-Active-Proxy (from jwilder/docker-gen)
		"exists":            exists,
		"groupByMulti":      groupByMulti,
		"dict":              dict,
		"trimSuffix":        strings.TrimSuffix,
		"closest":           arrayClosest,
		"first":             arrayFirst,
		"coalesce":          coalesce,
		"trim":              strings.TrimSpaces,    
	}
}

// serviceFunc returns a single service given a string argument in the form
// <service-name>[.<stack-name>].
func serviceFunc(ctx *TemplateContext) func(...string) (interface{}, error) {
	return func(s ...string) (result interface{}, err error) {
		result, err = ctx.GetService(s...)
		if _, ok := err.(NotFoundError); ok {
			log.Debug(err)
			return nil, nil
		}
		return
	}
}

// servicesFunc returns all available services, optionally filtered by stack
// name or label values.
func servicesFunc(ctx *TemplateContext) func(...string) (interface{}, error) {
	return func(s ...string) (interface{}, error) {
		return ctx.GetServices(s...)
	}
}

// hostFunc returns a single host given it's UUID.
func hostFunc(ctx *TemplateContext) func(...string) (interface{}, error) {
	return func(s ...string) (result interface{}, err error) {
		result, err = ctx.GetHost(s...)
		if _, ok := err.(NotFoundError); ok {
			log.Debug(err)
			return nil, nil
		}
		return
	}
}

// hostsFunc returns all available hosts, optionally filtered by label value.
func hostsFunc(ctx *TemplateContext) func(...string) (interface{}, error) {
	return func(s ...string) (interface{}, error) {
		return ctx.GetHosts(s...)
	}
}

// groupByLabel takes a label key and a slice of services or hosts and returns a map based
// on the values of the label.
//
// The map key is a string representing the label value. The map value is a
// slice of services or hosts that have the corresponding label value.
// Example:
//    {{range $labelValue, $containers := svc.Containers | groupByLabel "foo"}}
func groupByLabel(label string, in interface{}) (map[string][]interface{}, error) {
	m := make(map[string][]interface{})

	if in == nil {
		return m, fmt.Errorf("(groupByLabel) input is nil")
	}

	switch typed := in.(type) {
	case []Service:
		for _, s := range typed {
			value, ok := s.Labels[label]
			if ok && len(value) > 0 {
				m[value] = append(m[value], s)
			}
		}
	case []Container:
		for _, c := range typed {
			value, ok := c.Labels[label]
			if ok && len(value) > 0 {
				m[value] = append(m[value], c)
			}
		}
	case []Host:
		for _, h := range typed {
			value, ok := h.Labels[label]
			if ok && len(value) > 0 {
				m[value] = append(m[value], h)
			}
		}
	default:
		return m, fmt.Errorf("(groupByLabel) invalid input type %T", in)
	}

	return m, nil
}

//RAP : GroupbyMulti ( from jwilder/dockergen)
func groupByMulti(label string, sep string, in interface{}) (map[string][]interface{}, error) {
	m := make(map[string][]interface{})

	if in == nil {
		return m, fmt.Errorf("(groupByLabel) input is nil")
	}

	switch typed := in.(type) {
	case []Service:
		for _, s := range typed {
			value, ok := s.Labels[label]
			if ok && len(value) > 0 {
				items := strings.Split(string(value), sep)
				for _, item := range items {
				m[item] = append(m[item], s)
				}
			}
		}
	case []Container:
		for _, c := range typed {
			value, ok := c.Labels[label]
			if ok && len(value) > 0 {
				items := strings.Split(string(value), sep)
				for _, item := range items {
				m[item] = append(m[item], c)
				}
			}
		}
	case []Host:
		for _, h := range typed {
			value, ok := h.Labels[label]
			if ok && len(value) > 0 {
				items := strings.Split(string(value), sep)
				for _, item := range items {
				m[item] = append(m[item], h)
				}
			}
		}
	default:
		return m, fmt.Errorf("(groupByLabel) invalid input type %T", in)
	}

	return m, nil
}

func whereLabel(funcName string, in interface{}, label string, test func(string, bool) bool) ([]interface{}, error) {
	result := make([]interface{}, 0)
	if in == nil {
		return result, fmt.Errorf("(%s) input is nil", funcName)
	}
	if label == "" {
		return result, fmt.Errorf("(%s) label is empty", funcName)
	}

	switch typed := in.(type) {
	case []Service:
		for _, s := range typed {
			value, ok := s.Labels[label]
			if test(value, ok) {
				result = append(result, s)
			}
		}
	case []Container:
		for _, c := range typed {
			value, ok := c.Labels[label]
			if test(value, ok) {
				result = append(result, c)
			}
		}
	case []Host:
		for _, s := range typed {
			value, ok := s.Labels[label]
			if test(value, ok) {
				result = append(result, s)
			}
		}
	default:
		return result, fmt.Errorf("(%s) invalid input type %T", funcName, in)
	}

	return result, nil
}

// selects services or hosts from the input that have the given label
func whereLabelExists(label string, in interface{}) ([]interface{}, error) {
	return whereLabel("whereLabelExists", in, label, func(_ string, ok bool) bool {
		return ok
	})
}

// selects services or hosts from the input that have the given label and value
func whereLabelEquals(label, labelValue string, in interface{}) ([]interface{}, error) {
	return whereLabel("whereLabelEquals", in, label, func(value string, ok bool) bool {
		return ok && strings.EqualFold(value, labelValue)
	})
}

// selects services or hosts from the input that have the given label whose value matches the regex
func whereLabelMatches(label, pattern string, in interface{}) ([]interface{}, error) {
	rx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return whereLabel("whereLabelMatches", in, label, func(value string, ok bool) bool {
		return ok && rx.MatchString(value)
	})
}
