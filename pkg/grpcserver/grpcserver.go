package grpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strconv"

	"github.com/ansys/allie-flowkit/pkg/externalfunctions"
	"github.com/ansys/allie-sharedtypes/pkg/allieflowkitgrpc"
	"github.com/ansys/allie-sharedtypes/pkg/logging"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"google.golang.org/grpc"
)

// server is used to implement grpc_definition.ExternalFunctionsServer.
type server struct {
	allieflowkitgrpc.UnimplementedExternalFunctionsServer
}

// StartServer starts the gRPC server
// The server listens on the port specified in the configuration file
// The server implements the ExternalFunctionsServer interface
func StartServer() {
	lis, err := net.Listen("tcp", ":"+config.GlobalConfig.EXTERNALFUNCTIONS_GRPC_PORT)
	if err != nil {
		logging.Log.Fatalf(internalstates.Ctx, "failed to listen: %v", err)
	}
	s := grpc.NewServer()
	allieflowkitgrpc.RegisterExternalFunctionsServer(s, &server{})
	logging.Log.Infof(internalstates.Ctx, "gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		logging.Log.Fatalf(internalstates.Ctx, "failed to serve: %v", err)
	}
}

// ListFunctions lists all available function from the external functions package
//
// Parameters:
// - ctx: the context of the request
// - req: the request to list all available functions
//
// Returns:
// - allieflowkitgrpc.ListOfFunctions: a list of all available functions
// - error: an error if the function fails
func (s *server) ListFunctions(ctx context.Context, req *allieflowkitgrpc.ListFunctionsRequest) (*allieflowkitgrpc.ListFunctionsResponse, error) {

	// return all available functions
	return &allieflowkitgrpc.ListFunctionsResponse{Functions: internalstates.AvailableFunctions}, nil
}

// RunFunction runs a function from the external functions package
// The function is identified by the function id
// The function inputs are passed as a list of FunctionInput
//
// Parameters:
// - ctx: the context of the request
// - req: the request to run a function
//
// Returns:
// - allieflowkitgrpc.FunctionOutputs: the outputs of the function
// - error: an error if the function fails
func (s *server) RunFunction(ctx context.Context, req *allieflowkitgrpc.FunctionInputs) (output *allieflowkitgrpc.FunctionOutputs, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("error occured in gRPC server allie-flowkit during RunFunction of '%v': %v", req.Name, r)
		}
	}()

	// get function definition from available functions
	functionDefinition, ok := internalstates.AvailableFunctions[req.Name]
	if !ok {
		return nil, fmt.Errorf("function with name %s not found", req.Name)
	}

	// create input slice
	inputs := make([]interface{}, len(functionDefinition.Input))

	// unmarshal json string values for each input into the correct type
	for i, input := range req.Inputs {
		var err error
		inputs[i], err = convertStringToGivenType(input.Value, functionDefinition.Input[i].GoType)
		if err != nil {
			return nil, fmt.Errorf("error converting input %s to type %s: %v", input.Name, functionDefinition.Input[i].GoType, err)
		}

		// check for option sets and convert values
		if len(functionDefinition.Input[i].Options) > 0 {
			// convert value to correct type
			inputs[i], err = convertOptionSetValues(functionDefinition.Name, input.Name, inputs[i])
			if err != nil {
				return nil, fmt.Errorf("error converting input %s to type %s: %v", input.Name, functionDefinition.Input[i].GoType, err)
			}
		}
	}

	// get externalfunctions package and the function
	function, exists := externalfunctions.ExternalFunctionsMap[functionDefinition.Name]
	if !exists {
		return nil, fmt.Errorf("function %s not found in externalfunctions package", functionDefinition.Name)
	}
	funcValue := reflect.ValueOf(function)
	if !funcValue.IsValid() {
		return nil, fmt.Errorf("function %s not found in externalfunctions package", functionDefinition.Name)
	}

	// Prepare arguments for the function
	args := []reflect.Value{}
	for _, input := range inputs {
		args = append(args, reflect.ValueOf(input))
	}

	// Call the function
	results := funcValue.Call(args)

	// create output slice
	outputs := []*allieflowkitgrpc.FunctionOutput{}
	for i, result := range results {
		// marshal value to json string
		value, err := convertGivenTypeToString(result.Interface(), functionDefinition.Output[i].GoType)
		if err != nil {
			return nil, fmt.Errorf("error converting output %s to string: %v", functionDefinition.Output[i].Name, err)
		}

		// append output to slice
		outputs = append(outputs, &allieflowkitgrpc.FunctionOutput{
			Name:   functionDefinition.Output[i].Name,
			GoType: functionDefinition.Output[i].GoType,
			Value:  value,
		})
	}

	// return outputs
	return &allieflowkitgrpc.FunctionOutputs{Name: req.Name, Outputs: outputs}, nil
}

// StreamFunction streams a function from the external functions package
// The function is identified by the function id
// The function inputs are passed as a list of FunctionInput
//
// Parameters:
// - req: the request to stream a function
// - stream: the stream to send the function outputs
//
// Returns:
// - error: an error if the function fails
func (s *server) StreamFunction(req *allieflowkitgrpc.FunctionInputs, stream allieflowkitgrpc.ExternalFunctions_StreamFunctionServer) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("error occured in gRPC server allie-flowkit during StreamFunction of '%v': %v", req.Name, r)
		}
	}()

	// get function definition from available functions
	functionDefinition, ok := internalstates.AvailableFunctions[req.Name]
	if !ok {
		return fmt.Errorf("function with id %s not found", req.Name)
	}

	// create input slice
	inputs := make([]interface{}, len(functionDefinition.Input))

	// unmarshal json string values for each input into the correct type
	for i, input := range req.Inputs {
		var err error
		inputs[i], err = convertStringToGivenType(input.Value, functionDefinition.Input[i].GoType)
		if err != nil {
			return fmt.Errorf("error converting input %s to type %s: %v", input.Name, functionDefinition.Input[i].GoType, err)
		}

		// check for option sets and convert values
		if len(functionDefinition.Input[i].Options) > 0 {
			// convert value to correct type
			inputs[i], err = convertOptionSetValues(functionDefinition.Name, input.Name, inputs[i])
			if err != nil {
				return fmt.Errorf("error converting input %s to type %s: %v", input.Name, functionDefinition.Input[i].GoType, err)
			}
		}
	}

	// get externalfunctions package and the function
	function, exists := externalfunctions.ExternalFunctionsMap[functionDefinition.Name]
	if !exists {
		return fmt.Errorf("function %s not found in externalfunctions package", functionDefinition.Name)
	}
	funcValue := reflect.ValueOf(function)
	if !funcValue.IsValid() {
		return fmt.Errorf("function %s not found in externalfunctions package", functionDefinition.Name)
	}

	// Prepare arguments for the function
	args := []reflect.Value{}
	for _, input := range inputs {
		args = append(args, reflect.ValueOf(input))
	}

	// Call the function
	results := funcValue.Call(args)

	// get stream channel from results
	var streamChannel *chan string
	for i, output := range functionDefinition.Output {
		if output.GoType == "*chan string" {
			streamChannel = results[i].Interface().(*chan string)
		}
	}

	// listen to channel and send to stream
	var counter int32
	for message := range *streamChannel {
		// create output
		output := &allieflowkitgrpc.StreamOutput{
			MessageCounter: counter,
			IsLast:         false,
			Value:          message,
		}

		// send output to stream
		err := stream.Send(output)
		if err != nil {
			return err
		}

		// increment counter
		counter++
	}

	// send last message
	output := &allieflowkitgrpc.StreamOutput{
		MessageCounter: counter,
		IsLast:         true,
		Value:          "",
	}
	err = stream.Send(output)
	if err != nil {
		return err
	}

	return nil
}

// convertStringToGivenType converts a string to a given Go type.
//
// Parameters:
// - value: a string containing the value to convert
// - goType: a string containing the Go type to convert to
//
// Returns:
// - output: an interface containing the converted value
// - err: an error containing the error message
func convertStringToGivenType(value string, goType string) (output interface{}, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in convertStringToGivenType: %v", r)
		}
	}()

	switch goType {
	case "string":
		return value, nil
	case "float32":
		if value == "" {
			value = "0"
		}
		valueFloat64, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, err
		}
		return float32(valueFloat64), nil
	case "float64":
		if value == "" {
			value = "0"
		}
		return strconv.ParseFloat(value, 64)
	case "int":
		if value == "" {
			value = "0"
		}
		return strconv.Atoi(value)
	case "uint32":
		if value == "" {
			value = "0"
		}
		valueUint64, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, err
		}
		return uint32(valueUint64), nil
	case "bool":
		if value == "" {
			value = "false"
		}
		return strconv.ParseBool(value)
	case "[]string":
		if value == "" {
			value = "[]"
		}
		output := []string{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "[]float32":
		if value == "" {
			value = "[]"
		}
		output := []float32{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "[]float64":
		if value == "" {
			value = "[]"
		}
		output := []float64{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "[]int":
		if value == "" {
			value = "[]"
		}
		output := []int{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "[]bool":
		if value == "" {
			value = "[]"
		}
		output := []bool{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "map[string]string":
		if value == "" {
			value = "{}"
		}
		output := map[string]string{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "map[string]float64":
		if value == "" {
			value = "{}"
		}
		output := map[string]float64{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "map[string]int":
		if value == "" {
			value = "{}"
		}
		output := map[string]int{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "map[string]bool":
		if value == "" {
			value = "{}"
		}
		output := map[string]bool{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "map[string][]string":
		if value == "" {
			value = "{}"
		}
		output := map[string][]string{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "DbArrayFilter":
		if value == "" {
			value = "{}"
		}
		output := externalfunctions.DbArrayFilter{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "DbFilters":
		if value == "" {
			value = "{}"
		}
		output := externalfunctions.DbFilters{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "[]DbJsonFilter":
		if value == "" {
			value = "[]"
		}
		output := []externalfunctions.DbJsonFilter{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "[]DbResponse":
		if value == "" {
			value = "[]"
		}
		output := []externalfunctions.DbResponse{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "[]HistoricMessage":
		if value == "" {
			value = "[]"
		}
		output := []externalfunctions.HistoricMessage{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "*chan string":
		var output *chan string
		output = nil
		return output, nil

	case "[]AnsysGPTDefaultFields":
		if value == "" {
			value = "[]"
		}
		output := []externalfunctions.AnsysGPTDefaultFields{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil

	case "[]ACSSearchResponse":
		if value == "" {
			value = "[]"
		}
		output := []externalfunctions.ACSSearchResponse{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil

	case "[]AnsysGPTCitation":
		if value == "" {
			value = "[]"
		}
		output := []externalfunctions.AnsysGPTCitation{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil

	case "[]DataExtractionDocumentData":
		if value == "" {
			value = "[]"
		}
		output := []externalfunctions.DataExtractionDocumentData{}
		err := json.Unmarshal([]byte(value), &output)
		if err != nil {
			return nil, err
		}
		return output, nil
	}

	return nil, fmt.Errorf("unsupported GoType: '%s'", goType)
}

// convertGivenTypeToString converts a given Go type to a string.
//
// Parameters:
// - value: an interface containing the value to convert
// - goType: a string containing the Go type to convert from
//
// Returns:
// - string: a string containing the converted value
// - err: an error containing the error message
func convertGivenTypeToString(value interface{}, goType string) (output string, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("panic occured in ConvertGivenTypeToString: %v", r)
		}
	}()

	switch goType {
	case "string":
		return value.(string), nil
	case "float32":
		return strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32), nil
	case "float64":
		return strconv.FormatFloat(value.(float64), 'f', -1, 64), nil
	case "int":
		return strconv.Itoa(value.(int)), nil
	case "uint32":
		return strconv.FormatUint(uint64(value.(uint32)), 10), nil
	case "bool":
		return strconv.FormatBool(value.(bool)), nil
	case "[]string":
		output, err := json.Marshal(value.([]string))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]float32":
		output, err := json.Marshal(value.([]float32))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]float64":
		output, err := json.Marshal(value.([]float64))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]int":
		output, err := json.Marshal(value.([]int))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]bool":
		output, err := json.Marshal(value.([]bool))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "map[string]string":
		output, err := json.Marshal(value.(map[string]string))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "map[string]float64":
		output, err := json.Marshal(value.(map[string]float64))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "map[string]int":
		output, err := json.Marshal(value.(map[string]int))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "map[string]bool":
		output, err := json.Marshal(value.(map[string]bool))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "map[string][]string":
		output, err := json.Marshal(value.(map[string][]string))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "DbArrayFilter":
		output, err := json.Marshal(value.(externalfunctions.DbArrayFilter))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "DbFilters":
		output, err := json.Marshal(value.(externalfunctions.DbFilters))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]DbJsonFilter":
		output, err := json.Marshal(value.([]externalfunctions.DbJsonFilter))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]DbResponse":
		output, err := json.Marshal(value.([]externalfunctions.DbResponse))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]HistoricMessage":
		output, err := json.Marshal(value.([]externalfunctions.HistoricMessage))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "*chan string":
		return "", nil
	case "[]AnsysGPTDefaultFields":
		output, err := json.Marshal(value.([]externalfunctions.AnsysGPTDefaultFields))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]ACSSearchResponse":
		output, err := json.Marshal(value.([]externalfunctions.ACSSearchResponse))
		if err != nil {
			return "", err
		}
		return string(output), nil
	case "[]AnsysGPTCitation":
		output, err := json.Marshal(value.([]externalfunctions.AnsysGPTCitation))
		if err != nil {
			return "", err
		}
		return string(output), nil

	case "[]DataExtractionDocumentData":
		output, err := json.Marshal(value.([]externalfunctions.DataExtractionDocumentData))
		if err != nil {
			return "", err
		}
		return string(output), nil
	}

	return "", fmt.Errorf("unsupported GoType: '%s'", goType)
}

// convertOptionSetValues converts the option set values for the given function and input
//
// Parameters:
// - functionName: a string containing the function name
// - inputName: a string containing the input name
// - inputValue: an interface containing the input value
//
// Returns:
// - interface: an interface containing the converted value
// - error: an error containing the error message
func convertOptionSetValues(functionName string, inputName string, inputValue interface{}) (interface{}, error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(internalstates.Ctx, "Panic occured in convertOptionSetValues: %v", r)
		}
	}()

	switch functionName {

	case "AppendMessageHistory":

		switch inputName {

		case "role":
			return externalfunctions.AppendMessageHistoryRole(inputValue.(string)), nil

		default:
			return nil, fmt.Errorf("unsupported input for function %v: '%s'", functionName, inputName)
		}
	}

	return nil, fmt.Errorf("unsupported function: '%s'", functionName)
}
