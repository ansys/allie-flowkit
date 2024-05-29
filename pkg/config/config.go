package config

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

///////////////////////////////////////////////////////////////////////////////
// Types and global variables
///////////////////////////////////////////////////////////////////////////////

// Global configuration variable
var AllieFlowkitConfig *AllieFlowkitConfigStruct

// TODO: specify mandatory fields and optional fields

// ALLIEAgentConfigStruct is the structure of the configuration file
type AllieFlowkitConfigStruct struct {
	EXTERNALFUNCTIONS_GRPC_PORT *string `yaml:"EXTERNALFUNCTIONS_GRPC_PORT,omitempty"`
	LLM_HANDLER_ENDPOINT        *string `yaml:"LLM_HANDLER_ENDPOINT,omitempty"`
	KNOWLEDGE_DB_ENDPOINT       *string `yaml:"KNOWLEDGE_DB_ENDPOINT,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// Public Functions
///////////////////////////////////////////////////////////////////////////////

// LoadConfigFromFile loads the configuration from a YAML file into the global variable.
//
// This function is called from main() in main.go.
//
// Parameters:
//   - file: the YAML file to load the configuration from
//
// Returns:
//   - error: an error message if the YAML file cannot be loaded
func LoadConfigFromFile(file string) error {
	AllieFlowkitConfig = &AllieFlowkitConfigStruct{}
	_, err := AllieFlowkitConfig.ReadConfigFromFile(file)
	return err
}

// ReadConfigFromFile reads the configuration from a YAML file into the ALLIEAgentConfigStruct.
//
// Parameters:
//   - file: the YAML file to load the configuration from
//
// Returns:
//   - *ALLIEAgentConfigStruct: the configuration structure
//   - error: an error message if the YAML file cannot be loaded
func (c *AllieFlowkitConfigStruct) ReadConfigFromFile(file string) (*AllieFlowkitConfigStruct, error) {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		log.Printf("Error reading YAML file: %s\n", err)
		log.Println("Trying default YAML file location (configs/config.yaml)...")
		yamlFile, err = os.ReadFile("configs/config.yaml")
		if err != nil {
			log.Println("Default YAML file not found...")
			return nil, err
		}
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %s\n", err)
		return nil, err
	}

	// Check YAML entries missing
	missingFields := checkForMissingFields(c)

	if len(missingFields) > 0 {
		log.Fatalf("The following fields are missing from the YAML file: %v", missingFields)
	}

	log.Println("Configuration loaded successfully.")
	return c, nil
}

// ToString returns the configuration information in a string.
//
// Returns:
//   - string: the configuration information
func (c *AllieFlowkitConfigStruct) ToString() string {
	var info string
	v := reflect.ValueOf(*c)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldTag := t.Field(i).Tag
		fieldValue := v.Field(i)
		yamlKey := getYamlKeyFromTag(fieldTag)

		// Get the field value as a string
		// Fields are always pointers to the actual value
		var value string
		if fieldValue.IsNil() {
			value = "nil"
		} else {
			// Check if elem type before converting to string
			// If elem type is not a string, then we need to convert it to a string
			if fieldValue.Elem().Kind() == reflect.String {
				value = fieldValue.Elem().String()
			} else if fieldValue.Elem().Kind() == reflect.Bool {
				value = strconv.FormatBool(fieldValue.Elem().Bool())
			} else {
				log.Println("Unknown type for field:", yamlKey)
				continue
			}
		}

		info += yamlKey + ": " + value + "\n"
	}

	return info
}

///////////////////////////////////////////////////////////////////////////////
// Private Functions
///////////////////////////////////////////////////////////////////////////////

// checkForMissingFields checks for missing fields in the YAML file. It checks
// for missing fields by checking if the field is zero value (or nil).
//
// Parameters:
//   - c: the configuration structure
//
// Returns:
//   - []string: a list of missing fields
func checkForMissingFields(c *AllieFlowkitConfigStruct) []string {
	var missingFields []string
	v := reflect.ValueOf(*c)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldTag := t.Field(i).Tag
		fieldValue := v.Field(i)
		if isZero(fieldValue) {
			yamlKey := getYamlKeyFromTag(fieldTag)
			missingFields = append(missingFields, yamlKey)
		}
	}

	return missingFields
}

// isZero checks if a value is zero value (or nil).
//
// Parameters:
//   - v: the value to check
//
// Returns:
//   - bool: true if the value is zero value (or nil), false otherwise
func isZero(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// getYamlKeyFromTag gets the YAML key from the struct tag.
//
// Parameters:
//   - tag: the struct tag
//
// Returns:
//   - string: the YAML key
func getYamlKeyFromTag(tag reflect.StructTag) string {
	yamlTag := tag.Get("yaml")
	// Extract the key from the tag, removing the ",omitempty" flag
	parts := strings.Split(yamlTag, ",")
	return parts[0]
}
