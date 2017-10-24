// Package config get your application config values from different sources
// allowing you to run your app in different environments. It uses the go standard library
// to get the config values.
//
// Config properties are considered in the following order:
//
//   1 - Env variables (Upper case format, e.g. FIRST_NAME)
//   2 - config.json (Standard camel case syntax, e.g. firstName)
//
// Each item takes precedence over the item below it.

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/serenize/snaker"
)

// Property Object Definition.
type property struct {
	Field     reflect.Value
	FieldType reflect.Type
}

// Triggers when the config struct is not in the right type.
var ErrNotPointer = errors.New("invalid config object provided, should be a pointer")
var ErrNotStruct = errors.New("invalid config object provided, should be a struct")
var ErrNilStruct = errors.New("nil Struct pointer provided")

// ParseError occurs when an environment variable cannot be converted to
// the type required by a struct field during assignment.
type ParseError struct {
	FieldName    string
	TypeName     string
	ConfigSource string
	ConfigName   string
	ConfigValue  string
	Err          error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Error assigning %s (%s) to %s field: converting '%s' to type %s. Reason: %s", e.ConfigName, e.ConfigSource, e.FieldName, e.ConfigValue, e.TypeName, e.Err)
}

// Load parses your configuration values into the provided struct based on its fields and tags.
func Load(config interface{}) error {
	// shouldn't be nil
	if config == nil {
		return ErrNilStruct
	}

	// Getting reflection object value
	v := reflect.ValueOf(config)

	// Config has to be a pointer
	if v.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	// Config should not be nil
	if v.IsNil() {
		return ErrNilStruct
	}

	// Config has to be a struct
	if reflect.Indirect(v).Kind() != reflect.Struct {
		return ErrNotStruct
	}

	// Loading config.json config first if its exists.
	if err := LoadJSONConfig(config); err != nil {
		return err
	}

	// Override loaded values with env and flag options
	if err := loadStruct(v, ""); err != nil {
		return err
	}

	return nil
}

// LoadJSONConfig loads config from json file. It uses
// the json.Unmarshal function for loading the json to the struct.
func LoadJSONConfig(config interface{}) error {

	// Checking if the config file env variable exists
	envName := getEnvConfigName("configJSONFile", "")
	if envValue, found := os.LookupEnv(envName); found {
		// Checking if the file exists
		if _, err := os.Stat(envValue); err == nil {
			file, e := ioutil.ReadFile(envValue)
			if e != nil {
				return err
			}

			return json.Unmarshal(file, &config)
		}

	}

	// Checking if the `config.json` file exists
	if _, err := os.Stat("config.json"); err == nil {
		file, e := ioutil.ReadFile("config.json")
		if e != nil {
			return err
		}

		return json.Unmarshal(file, &config)
	}

	return nil
}

// loadStruct loop over the struct's fields and
// match the field name with the config name. This
// function might be called recursively for nested struct.
func loadStruct(v reflect.Value, prefix string) error {
	s := v

	// if a pointer to a struct is passed, get the type of the referenced object
	if s.Kind() == reflect.Ptr {
		s = v.Elem()
	}

	// Iterate over the struct's fields
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		t := s.Type().Field(i)

		// Checking if this field can be set.
		if !f.CanSet() {
			// Ignore and continue
			continue
		}

		// Checking if this field was masked as ignored
		if _, ok := t.Tag.Lookup("ignored"); ok {
			// Ignore and continue
			continue
		}

		// Checking if its a pointer to struct.
		if f.Kind() == reflect.Ptr {
			// Checking if we need to initialize the struct.
			if f.IsNil() {
				f.Set(reflect.New(f.Type().Elem()))
			}

			f = f.Elem()
		}

		// If field is a struct, we will loop over its fields recursively
		if f.Kind() == reflect.Struct {
			if err := loadStruct(f, getConfigNamePrefix(t.Name, prefix)); err != nil {
				return err
			}
		}

		// This is a regular field so lets try to parse it.
		if err := loadProperty(f, t, prefix); err != nil {
			return err
		}

	}

	return nil
}

// loadProperty gets the value from env or flag and assign it to the field.
// Only integer, float, string and boolean are supported.
func loadProperty(v reflect.Value, t reflect.StructField, prefix string) error {
	fieldName := t.Name

	// Checking if this field has an alias
	if value, ok := t.Tag.Lookup("alias"); ok {
		fieldName = value
	}

	// Getting config from env
	envName := getEnvConfigName(fieldName, prefix)
	envValue, found := os.LookupEnv(envName)
	configSource := "Environment"

	if !found {
		// Check if this field has a default value
		if value, ok := t.Tag.Lookup("default"); ok {
			envValue = value
			// Check if the field is mandatory, if it is return an error.
		} else if value, ok := t.Tag.Lookup("required"); ok && value == "true" {
			return &ParseError{
				ConfigValue:  envValue,
				ConfigName:   envName,
				ConfigSource: configSource,
				FieldName:    fieldName,
				TypeName:     v.Type().Name(),
				Err:          errors.New("This field is required"),
			}
		} else {
			// Ignore property and continue
			return nil
		}
	}

	// Setting values
	switch v.Kind() {
	case reflect.String:
		v.SetString(envValue)

	case reflect.Bool:
		val, err := strconv.ParseBool(envValue)
		if err != nil {
			return &ParseError{
				ConfigValue:  envValue,
				ConfigName:   envName,
				ConfigSource: configSource,
				FieldName:    fieldName,
				TypeName:     v.Type().Name(),
				Err:          err,
			}
		}
		v.SetBool(val)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(envValue, 0, v.Type().Bits())
		if err != nil {
			return &ParseError{
				ConfigValue:  envValue,
				ConfigName:   envName,
				ConfigSource: configSource,
				FieldName:    fieldName,
				TypeName:     v.Type().Name(),
				Err:          err,
			}
		}

		v.SetInt(val)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(envValue, 0, v.Type().Bits())
		if err != nil {
			return &ParseError{
				ConfigValue:  envValue,
				ConfigName:   envName,
				ConfigSource: configSource,
				FieldName:    fieldName,
				TypeName:     v.Type().Name(),
				Err:          err,
			}
		}
		v.SetUint(val)

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(envValue, v.Type().Bits())
		if err != nil {
			return &ParseError{
				ConfigValue:  envValue,
				ConfigName:   envName,
				ConfigSource: configSource,
				FieldName:    fieldName,
				TypeName:     v.Type().Name(),
				Err:          err,
			}
		}
		v.SetFloat(val)

	default:
		return &ParseError{
			ConfigValue:  envValue,
			ConfigName:   envName,
			ConfigSource: configSource,
			FieldName:    fieldName,
			TypeName:     v.Type().Name(),
			Err:          fmt.Errorf("unsupported variable type"),
		}
	}

	return nil
}

// getConfigNamePrefix returns the config name prefix formatted as <prefix>_<field name>.
// IF not prefix is provided, the field name will be return.
func getConfigNamePrefix(fieldName, prefix string) string {
	if prefix != "" {
		return fmt.Sprintf("%s_%s", strings.TrimSuffix(prefix, "_"), fieldName)
	}

	return fieldName
}

// getEnvConfigName returns the name of the environment variable in UPPER_SNAKE_CASE.
// Field name should be in camelcase.
func getEnvConfigName(fieldName, prefix string) string {
	var envName string
	if prefix != "" {
		envName = fmt.Sprintf("%s_%s", snaker.CamelToSnake(strings.TrimSuffix(prefix, "_")), snaker.CamelToSnake(fieldName))
	} else {
		envName = snaker.CamelToSnake(fieldName)
	}

	return strings.ToUpper(envName)
}
