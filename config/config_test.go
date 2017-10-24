package config

import (
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

// Test the config name prefix
func TestPrefixConfigName(t *testing.T) {
	tt := []struct {
		name   string
		prefix string
		result string
	}{
		{"", "", ""},
		{"Debug", "", "Debug"},
		{"Username", "Db", "Db_Username"},
		{"age", "person", "person_age"},
		{"City", "Address_", "Address_City"},
		{"user_name", "db", "db_user_name"},
	}

	for _, tc := range tt {
		result := getConfigNamePrefix(tc.name, tc.prefix)

		if result != tc.result {
			t.Errorf("Invalid prefix [name='%s'] [prefix='%s] - Expected: %s; Got: %s", tc.name, tc.prefix, tc.result, result)
		}
	}
}

// Test the env config name. Enviroment variable name should
// be in uppercase snake_case.
func TestEnvConfigName(t *testing.T) {
	tt := []struct {
		name   string
		prefix string
		result string
	}{
		{"", "", ""},
		{"Debug", "", "DEBUG"},
		{"Port", "", "PORT"},
		{"Username", "DB", "DB_USERNAME"},
		{"age", "person", "PERSON_AGE"},
		{"City", "Address_", "ADDRESS_CITY"},
		{"user_name", "db", "DB_USER_NAME"},
		{"personAddress", "", "PERSON_ADDRESS"},
		{"phoneNumber", "User", "USER_PHONE_NUMBER"},
		{"fax#", "User", "USER_FAX#"},
		{"Integer1", "", "INTEGER1"},
	}

	for _, tc := range tt {
		result := getEnvConfigName(tc.name, tc.prefix)

		if result != tc.result {
			t.Errorf("Invalid env config name [name='%s'] [prefix='%s] - Expected: %s; Got: %s", tc.name, tc.prefix, tc.result, result)
		}
	}
}

// Testing a non-pointer struct.
func TestLoadNonPointerConfig(t *testing.T) {
	type TestConfig struct {
	}

	var c TestConfig

	// The config struct have to be a pointer so it can be populated
	err := Load(c)
	if err != nil && err != ErrNotPointer {
		t.Errorf("Non pointer Struct - Expected: %+v; Got: %+v", ErrNotPointer, err)
	}
}

// Testing nil struct.
func TestLoadNilConfig(t *testing.T) {
	type TestConfig struct {
	}

	var c *TestConfig

	// The config struct can't be nil
	err := Load(c)
	if err != nil && err != ErrNilStruct {
		t.Errorf("Nil Struct Pointer - Expecting: %+v; Got: %+v", ErrNilStruct, err)
	}
}

// Testing struct without properties
func TestLoadConfigWithoutElement(t *testing.T) {
	type TestConfig struct {
	}

	c := &TestConfig{}

	// This shouldn't fail
	err := Load(c)
	if err != nil {
		t.Errorf("Struct Without Element - Expecting: nil; Got: %+v", err)
	}
}

// Testing nil
func TestNil(t *testing.T) {
	err := Load(nil)
	if err != ErrNilStruct {
		t.Errorf("Struct Without Element - Expecting: %v; Got: %v", ErrNilStruct, err)
	}
}

// Testing a non struct.
func TestInvalidObject(t *testing.T) {
	// Using interface
	type TestConfig interface {
	}

	var c TestConfig

	tt := []struct {
		name   string
		config interface{}
		err    error
	}{
		{"Interface", c, ErrNotStruct},
		{"Map", map[string]string{}, ErrNotStruct},
		{"Slice", []string{}, ErrNotStruct},
		{"Func", func(key, value string) {}, ErrNotStruct},
		{"String", "Test", ErrNotStruct},
		{"Boolean", true, ErrNotStruct},
		{"Int", 1, ErrNotStruct},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := Load(&tc.config)
			if err != tc.err {
				t.Errorf("%s - Expecting: %v; Got: %+v", tc.name, tc.err, err)
			}
		})
	}
}

// Testing unsetable fields
func TestUnsetableField(t *testing.T) {
	c := &struct {
		username string
	}{}
	os.Setenv("USERNAME", "gohan")
	err := Load(c)
	if err != nil {
		t.Errorf("Unsetable field - Expecting: nil; Got: %v", err)
	}

	if c.username != "" {
		t.Errorf("Unsetable field - Expecting: empty; Got: %v", c.username)
	}

	os.Unsetenv("USERNAME")
}

// Testing nested struct pointer
func TestNestedStructPointer(t *testing.T) {
	c := &struct {
		// Struct poinger
		Db *struct {
			Username   string
			AutoCommit bool
		}
		// Struct
		Db2 struct {
			Username string
		}
	}{}
	os.Setenv("DB_USERNAME", "gohan")
	flag.Set("db-username", "test-user")
	os.Setenv("DB2_USERNAME", "gohan2")
	err := Load(c)
	if err != nil {
		t.Errorf("Nested struct pointer - Expecting: nil; Got: %v", err)
	}

	if c.Db.Username != "gohan" {
		t.Errorf("Nested struct pointer - Expecting: gohan; Got: %v", c.Db.Username)
	}

	if c.Db2.Username != "gohan2" {
		t.Errorf("Nested struct pointer - Expecting: gohan; Got: %v", c.Db.Username)
	}

	os.Unsetenv("DB_USERNAME")

	// Testing nested struct parsing error
	os.Setenv("DB_AUTO_COMMIT", "test")
	err = Load(c)
	if err == nil {
		t.Errorf("Nested struct pointer - Expecting: ParseError, Got: nil")
	}

	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("Nested struct pointer - Expecting: ParseError, Got %v", v)
	}
	if v.FieldName != "AutoCommit" {
		t.Errorf("Nested struct pointer - Expecting: AutoCommit, Got %v", v.FieldName)
	}
	// Should be false because it couldn't be assigned
	if c.Db.AutoCommit != false {
		t.Errorf("Nested struct pointer - Expecting: %v, Got: %v", false, c.Db.AutoCommit)
	}

	os.Unsetenv("DB_AUTO_COMMIT")

}

// Testing ignored fields
func TestLoadIgnoreField(t *testing.T) {
	c := &struct {
		Username string `ignored:"true"`
		Debug    bool   `ignored:"true"`
	}{
		Debug: true,
	}

	os.Setenv("USERNAME", "gohan")
	os.Setenv("DEBUG", "false")
	err := Load(c)
	if err != nil {
		t.Errorf("Ignored Field - Expecting: no error, Got: %v", err)
	}

	if c.Username != "" {
		t.Errorf("Required Field - Expecting: '', Got: %+v", c.Username)
	}

	if !c.Debug {
		t.Errorf("Required Field - Expecting: true, Got: %+v", c.Debug)
	}

	os.Unsetenv("USERNAME")
	os.Unsetenv("DEBUG")

}

// Testing required fields
func TestLoadRequiredField(t *testing.T) {
	c := &struct {
		Username string `required:"true"`
	}{}

	os.Unsetenv("USERNAME")
	err := Load(c)
	if err == nil {
		t.Errorf("Required Field - Expecting: error, Got: nil")
	}

	// Trying to convert the error to Parse error.
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("Required Field - Expecting: ParseError, Got %v", v)
	}

	os.Setenv("USERNAME", "gohan")
	err = Load(c)
	if err != nil {
		t.Errorf("Required Field - Expecting: nil, Got: %+v", err)
	}

	if c.Username != "gohan" {
		t.Errorf("Required Field - Expecting: gohan, Got: %+v", c.Username)
	}

}

// Testing default field value
func TestLoadDefaultFieldValue(t *testing.T) {
	c := &struct {
		Username string `default:"gohan"`
		Debug    bool   `default:"true"`
	}{}

	os.Unsetenv("USERNAME")
	os.Unsetenv("DEBUG")

	err := Load(c)
	if err != nil {
		t.Errorf("Ignored Field - Expecting: no error, Got: %v", err)
	}

	if c.Username != "gohan" {
		t.Errorf("Required Field - Expecting: gohan, Got: %+v", c.Username)
	}

	if !c.Debug {
		t.Errorf("Required Field - Expecting: true, Got: %+v", c.Debug)
	}

}

// Testing Field's Alias
func TestAliasFieldLoading(t *testing.T) {
	c := &struct {
		Username string `alias:"UserName"`
		Debug    bool   `alias:"log_debug"`
	}{}

	os.Setenv("USER_NAME", "gohan")
	os.Setenv("LOG_DEBUG", "true")
	err := Load(c)
	if err != nil {
		t.Errorf("Field Alias - Expecting: no error, Got: %v", err)
	}

	if c.Username != "gohan" {
		t.Errorf("Field Alias - Expecting: 'gohan', Got: %+v", c.Username)
	}

	if !c.Debug {
		t.Errorf("Field Alias - Expecting: true, Got: %+v", c.Debug)
	}

}

// Testing boolean values loading. The config packages uses
// the strconv.ParseBool for converting the env/flag string
// values to bool. Note that this functions accepts 1, t, T, TRUE, true, True,
// 0, f, F, FALSE, false and  False values
func TestLoadBoolConfigValue(t *testing.T) {

	type TestConfig struct {
		Debug bool
		Db    struct {
			AutoCommit bool
		}
	}

	tt := []struct {
		name                string
		envName             string
		config              *TestConfig
		envValue            string
		result              bool
		shourlReturnError   bool
		errShouldBeParseErr bool
	}{
		{"Should be true (Env Value = true)", "DEBUG", &TestConfig{}, "true", true, false, false},
		{"Should be true (Env Value = TRUE)", "DEBUG", &TestConfig{}, "TRUE", true, false, false},
		{"Should be true (Env Value = True)", "DEBUG", &TestConfig{}, "True", true, false, false},
		{"Should be true (Env Value = 1)", "DEBUG", &TestConfig{}, "1", true, false, false},
		{"Should be true (Env Value = T)", "DEBUG", &TestConfig{}, "T", true, false, false},
		{"Should be true (Env Value = t)", "DEBUG", &TestConfig{}, "t", true, false, false},

		{"Should be false (Env Value = false)", "DEBUG", &TestConfig{Debug: true}, "false", false, false, false},
		{"Should be false (Env Value = FALSE)", "DEBUG", &TestConfig{Debug: true}, "FALSE", false, false, false},
		{"Should be false (Env Value = False)", "DEBUG", &TestConfig{Debug: true}, "False", false, false, false},
		{"Should be false (Env Value = 0)", "DEBUG", &TestConfig{Debug: true}, "0", false, false, false},
		{"Should be false (Env Value = F)", "DEBUG", &TestConfig{Debug: true}, "F", false, false, false},
		{"Should be false (Env Value = f)", "DEBUG", &TestConfig{Debug: true}, "f", false, false, false},

		{"Should return error (Env Value = test)", "DEBUG", &TestConfig{}, "test", false, true, true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv(tc.envName, tc.envValue)
			err := Load(tc.config)

			if err != nil && !tc.shourlReturnError {
				t.Fatalf("%s - Expecting: no error; Got: %+v", tc.name, err)
			} else if err != nil && tc.errShouldBeParseErr {
				_, ok := err.(*ParseError)
				if !ok {
					t.Fatalf("%s - Expecting: no error; Got: %+v", tc.name, reflect.TypeOf(err))
				}
			}

			if tc.config.Debug != tc.result {
				t.Fatalf("%s - Expecting: %v; Got: %v", tc.name, tc.result, tc.config.Debug)
			}
		})
	}
}

// Testing String values loading.
func TestLoadStringConfigValue(t *testing.T) {
	c := &struct {
		Username    string
		PtrUsername *string
	}{}

	// Testing true value
	os.Setenv("USERNAME", "gohan")
	os.Setenv("PTR_USERNAME", "gohan2")
	err := Load(c)
	if err != nil {
		t.Errorf("Config String - Expecting: nil, Got: %+v", err)
	}

	if c.Username != "gohan" {
		t.Errorf("Config String - Expecting: gohan, Got: %s", c.Username)
	}

	if *c.PtrUsername != "gohan2" {
		t.Errorf("Config String - Expecting: gohan2, Got: %s", *c.PtrUsername)
	}

	os.Unsetenv("USERNAME")
	os.Unsetenv("PTR_USERNAME")

}

// Testing integer values loading.
func TestLoadIntegerConfigValue(t *testing.T) {

	c := &struct {
		Integer8  int8
		Integer16 int16
		Integer32 int32
		Integer64 int64

		PtrInteger8  *int8
		PtrInteger16 *int16
		PtrInteger32 *int32
		PtrInteger64 *int64
	}{}

	// [START int8]
	os.Setenv("INTEGER8", "10")
	os.Setenv("PTR_INTEGER8", "20")

	err := Load(c)
	if err != nil {
		t.Errorf("Config int8 - Expecting: nil, Got: %+v", err)
	}

	if c.Integer8 != 10 {
		t.Errorf("Config int8 - Expecting: 10, Got: %d", c.Integer8)
	}

	if *c.PtrInteger8 != 20 {
		t.Errorf("Config *int8 - Expecting: 20, Got: %d", *c.PtrInteger8)
	}
	// [END int8]

	// [START int16]
	os.Setenv("INTEGER16", "15")
	os.Setenv("PTR_INTEGER16", "16")

	err = Load(c)
	if err != nil {
		t.Errorf("Config int16 - Expecting: nil, Got: %+v", err)
	}

	if c.Integer16 != 15 {
		t.Errorf("Config int16 - Expecting: 15, Got: %d", c.Integer16)
	}

	if *c.PtrInteger16 != 16 {
		t.Errorf("Config *int16 - Expecting: 16, Got: %d", *c.PtrInteger16)
	}
	// [END int16]

	// [START int32]
	os.Setenv("INTEGER32", "12345")
	os.Setenv("PTR_INTEGER32", "54321")

	err = Load(c)
	if err != nil {
		t.Errorf("Config int32 - Expecting: nil, Got: %+v", err)
	}

	if c.Integer32 != 12345 {
		t.Errorf("Config int32 - Expecting: 12345 Got: %d", c.Integer32)
	}

	if *c.PtrInteger32 != 54321 {
		t.Errorf("Config *int32 - Expecting: 54321, Got: %d", *c.PtrInteger32)
	}
	// [END int32]

	// [START int64]
	os.Setenv("INTEGER64", "123456789")
	os.Setenv("PTR_INTEGER64", "987654321")

	err = Load(c)
	if err != nil {
		t.Errorf("Config int64 - Expecting: nil, Got: %+v", err)
	}

	if c.Integer64 != 123456789 {
		t.Errorf("Config int64 - Expecting: 123456789, Got: %d", c.Integer64)
	}

	if *c.PtrInteger64 != 987654321 {
		t.Errorf("Config *int64 - Expecting: 987654321, Got: %d", *c.PtrInteger64)
	}
	// [END int32]

	// Testing invalid int value
	os.Setenv("INTEGER32", "should be int32")
	err = Load(c)
	if err == nil {
		t.Errorf("Config int - Expecting: error, Got: nil")
	} else {
		// Trying to convert the error to ParseError.
		v, ok := err.(*ParseError)
		if !ok {
			t.Errorf("Config int - Expecting: ParseError, Got %v", v)
		}
	}
}

// Testing unsigned integer values loading.
func TestLoadUnsignedIntegerConfigValue(t *testing.T) {

	c := &struct {
		UnsignedInteger8  uint8
		UnsignedInteger16 uint16
		UnsignedInteger32 uint32
		UnsignedInteger64 uint64

		PtrUnsignedInteger8  *uint8
		PtrUnsignedInteger16 *uint16
		PtrUnsignedInteger32 *uint32
		PtrUnsignedInteger64 *uint64
	}{}

	// [START uint8]
	os.Setenv("UNSIGNED_INTEGER8", "10")
	os.Setenv("PTR_UNSIGNED_INTEGER8", "20")

	err := Load(c)
	if err != nil {
		t.Errorf("Config uint8 - Expecting: nil, Got: %+v", err)
	}

	if c.UnsignedInteger8 != 10 {
		t.Errorf("Config uint8 - Expecting: 10, Got: %d", c.UnsignedInteger8)
	}

	if *c.PtrUnsignedInteger8 != 20 {
		t.Errorf("Config *uint8 - Expecting: 20, Got: %d", *c.PtrUnsignedInteger8)
	}
	// [END uint8]

	// [START uint16]
	os.Setenv("UNSIGNED_INTEGER16", "15")
	os.Setenv("PTR_UNSIGNED_INTEGER16", "16")

	err = Load(c)
	if err != nil {
		t.Errorf("Config uint16 - Expecting: nil, Got: %+v", err)
	}

	if c.UnsignedInteger16 != 15 {
		t.Errorf("Config uint16 - Expecting: 15, Got: %d", c.UnsignedInteger16)
	}

	if *c.PtrUnsignedInteger16 != 16 {
		t.Errorf("Config *uint16 - Expecting: 16, Got: %d", *c.PtrUnsignedInteger16)
	}
	// [END uint16]

	// [START uint32]
	os.Setenv("UNSIGNED_INTEGER32", "12345")
	os.Setenv("PTR_UNSIGNED_INTEGER32", "54321")

	err = Load(c)
	if err != nil {
		t.Errorf("Config uint32 - Expecting: nil, Got: %+v", err)
	}

	if c.UnsignedInteger32 != 12345 {
		t.Errorf("Config uint32 - Expecting: 12345 Got: %d", c.UnsignedInteger32)
	}

	if *c.PtrUnsignedInteger32 != 54321 {
		t.Errorf("Config *uint32 - Expecting: 54321, Got: %d", *c.PtrUnsignedInteger32)
	}
	// [END uint32]

	// [START uint64]
	os.Setenv("UNSIGNED_INTEGER64", "123456789")
	os.Setenv("PTR_UNSIGNED_INTEGER64", "987654321")

	err = Load(c)
	if err != nil {
		t.Errorf("Config uint64 - Expecting: nil, Got: %+v", err)
	}

	if c.UnsignedInteger64 != 123456789 {
		t.Errorf("Config uint64 - Expecting: 123456789, Got: %d", c.UnsignedInteger64)
	}

	if *c.PtrUnsignedInteger64 != 987654321 {
		t.Errorf("Config *uint64 - Expecting: 987654321, Got: %d", *c.PtrUnsignedInteger64)
	}
	// [END int32]

	// Testing invalid int value
	os.Setenv("UNSIGNED_INTEGER32", "should be uint32")
	err = Load(c)
	if err == nil {
		t.Errorf("Config uint - Expecting: error, Got: nil")
	} else {
		// Trying to convert the error to Parse error.
		v, ok := err.(*ParseError)
		if !ok {
			t.Errorf("Config uint - Expecting: ParseError, Got %v", v)
		}
	}
}

// Testing float values loading.
func TestLoadFloatConfigValue(t *testing.T) {

	c := &struct {
		Float32 float32
		Float64 float64

		PtrFloat32 *float32
		PtrFloat64 *float64
	}{}

	// [START float32]
	os.Setenv("FLOAT32", "123.45")
	os.Setenv("PTR_FLOAT32", "543.21")

	err := Load(c)
	if err != nil {
		t.Errorf("Config float32 - Expecting: nil, Got: %+v", err)
	}

	if c.Float32 != 123.45 {
		t.Errorf("Config float32 - Expecting: 123.45 Got: %f", c.Float32)
	}

	if *c.PtrFloat32 != 543.21 {
		t.Errorf("Config *float32 - Expecting: 543.21, Got: %f", *c.PtrFloat32)
	}
	// [END float32]

	// [START float64]
	os.Setenv("FLOAT64", "1234567.89")
	os.Setenv("PTR_FLOAT64", "9876543.21")

	err = Load(c)
	if err != nil {
		t.Errorf("Config float64 - Expecting: nil, Got: %+v", err)
	}

	if c.Float64 != 1234567.89 {
		t.Errorf("Config float64 - Expecting: 1234567.89, Got: %f", c.Float64)
	}

	if *c.PtrFloat64 != 9876543.21 {
		t.Errorf("Config *float64 - Expecting: 9876543.21, Got: %f", *c.PtrFloat64)
	}
	// [END int32]

	// Testing invalid int value
	os.Setenv("FLOAT32", "should be float32")
	err = Load(c)
	if err == nil {
		t.Errorf("Config float32 - Expecting: error, Got: nil")
	} else {
		// Trying to convert the error to ParseError.
		v, ok := err.(*ParseError)
		if !ok {
			t.Errorf("Config float32 - Expecting: ParseError, Got %v", v)
		}
	}
}

// Testing float values loading.
func TestUnSupportedFieldTypeLoading(t *testing.T) {

	c := &struct {
		Roles []string
	}{}

	os.Setenv("ROLES", "STORE,ADMIN")

	err := Load(c)
	if err == nil {
		t.Errorf("Config Unsupported Type - Expecting: error, Got: nil")
	} else {
		// Trying to convert the error to ParseError.
		v, ok := err.(*ParseError)
		if !ok {
			t.Errorf("Config float32 - Expecting: ParseError, Got: %+v", v)
		}
	}
}

// Testing load json config file
func TestLoadJSONConfig(t *testing.T) {
	c := &struct {
		Username string
		Role     string

		Db struct {
			Host string
			Port int
		}
	}{}

	// [START json loading from env]
	json := `
	{
		"username" : "gohan",
		"role" : "admin",
		"db" : {
			"host": "localhost",
			"port": 3306
		}
	}
	`

	tmpfile, err := ioutil.TempFile("", "config-test")
	if err != nil {
		t.Errorf("Json Loading - Expecting: No error; Got: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write([]byte(json)); err != nil {
		t.Errorf("Json Loading - Expecting: No error; Got: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Errorf("Json Loading - Expecting: No error; Got: %v", err)
	}

	os.Setenv("CONFIG_JSON_FILE", tmpfile.Name())

	err = Load(c)
	if err != nil {
		t.Errorf("Json Loading - Expecting: No error; Got: %v", err)
	}

	if c.Username != "gohan" {
		t.Errorf("Json Loading - Expecting: gohan; Got: %s", c.Username)
	}

	if c.Role != "admin" {
		t.Errorf("Json Loading - Expecting: admin; Got: %s", c.Role)
	}

	if c.Db.Host != "localhost" {
		t.Errorf("Json Loading - Expecting: localhost; Got: %s", c.Db.Host)
	}

	if c.Db.Port != 3306 {
		t.Errorf("Json Loading - Expecting: 3306; Got: %d", c.Db.Port)
	}
	// [END json loading from env]

	// [START json loading from current dir]
	os.Unsetenv("CONFIG_JSON_FILE")
	json = `
	{
		"username" : "gohan2",
		"role" : "admin2",
		"db" : {
			"host": "localhost2",
			"port": 3307
		}
	}
	`
	if err := ioutil.WriteFile("config.json", []byte(json), os.ModePerm); err != nil {
		t.Errorf("Json Loading - Expecting: No error; Got: %v", err)
	}
	defer os.Remove("config.json")

	err = Load(c)
	if err != nil {
		t.Errorf("Json Loading - Expecting: No error; Got: %v", err)
	}

	if c.Username != "gohan2" {
		t.Errorf("Json Loading - Expecting: gohan2; Got: %s", c.Username)
	}

	if c.Role != "admin2" {
		t.Errorf("Json Loading - Expecting: admin; Got: %s", c.Role)
	}

	if c.Db.Host != "localhost2" {
		t.Errorf("Json Loading - Expecting: localhost; Got: %s", c.Db.Host)
	}

	if c.Db.Port != 3307 {
		t.Errorf("Json Loading - Expecting: 3306; Got: %d", c.Db.Port)
	}
	// [END json loading from current dir]

	// [START fail json loading]
	json = `
	{
		"username" : "gohan",
		"role" : "admin",
		"db" : {
			"host": "localhost",
			"port": 3306
		}
	` // With syntax error (missing } at the end)

	os.Remove("config.json")
	if err := ioutil.WriteFile("config.json", []byte(json), os.ModePerm); err != nil {
		t.Errorf("Json Loading - Expecting: No error; Got: %v", err)
	}
	defer os.Remove("config.json")

	err = Load(c)
	if err == nil {
		t.Errorf("Json Loading - Expecting: error; Got: nil")
	}

	// [END fail json loading]

}
