package grpcserver

import (
	"context"
	"fmt"
	"net"
	"reflect"

	"github.com/ansys/allie-flowkit/pkg/externalfunctions"
	"github.com/ansys/allie-sharedtypes/pkg/allieflowkitgrpc"
	"github.com/ansys/allie-sharedtypes/pkg/logging"
	"github.com/ansys/allie-sharedtypes/pkg/typeconverters"

	"github.com/ansys/allie-flowkit/pkg/internalstates"
	"github.com/ansys/allie-sharedtypes/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// server is used to implement grpc_definition.ExternalFunctionsServer.
type server struct {
	allieflowkitgrpc.UnimplementedExternalFunctionsServer
}

// StartServer starts the gRPC server
// The server listens on the port specified in the configuration file
// The server implements the ExternalFunctionsServer interface
func StartServer() {
	// Create listener on the specified port
	lis, err := net.Listen("tcp", ":"+config.GlobalConfig.EXTERNALFUNCTIONS_GRPC_PORT)
	if err != nil {
		logging.Log.Fatalf(internalstates.Ctx, "failed to listen: %v", err)
	}

	// Check if SSL is enabled and load the server's certificate and private key
	var opts []grpc.ServerOption
	if config.GlobalConfig.USE_SSL {
		creds, err := credentials.NewServerTLSFromFile(
			config.GlobalConfig.SSL_CERT_PUBLIC_KEY_FILE,
			config.GlobalConfig.SSL_CERT_PRIVATE_KEY_FILE,
		)
		if err != nil {
			logging.Log.Fatalf(internalstates.Ctx, "failed to load SSL certificates: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Add API key authentication interceptor if an API key is provided
	if config.GlobalConfig.FLOWKIT_API_KEY != "" {
		opts = append(opts, grpc.UnaryInterceptor(apiKeyAuthInterceptor(config.GlobalConfig.FLOWKIT_API_KEY)))
	}

	// Set gRPC message size limits
	opts = append(opts, grpc.MaxRecvMsgSize(1024*1024*1024)) // 1 GB receive limit
	opts = append(opts, grpc.MaxSendMsgSize(1024*1024*1024)) // 1 GB send limit

	// Create the gRPC server with the options
	s := grpc.NewServer(opts...)
	allieflowkitgrpc.RegisterExternalFunctionsServer(s, &server{})
	logging.Log.Infof(internalstates.Ctx, "gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		logging.Log.Fatalf(internalstates.Ctx, "failed to serve: %v", err)
	}
}

// apiKeyAuthInterceptor is a gRPC server interceptor that checks for a valid API key in the metadata of the request
// The API key is passed as a string parameter
//
// Parameters:
// - apiKey: a string containing the API key
//
// Returns:
// - grpc.UnaryServerInterceptor: a gRPC server interceptor
func apiKeyAuthInterceptor(apiKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract API key from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		receivedApiKeys := md["x-api-key"]
		if len(receivedApiKeys) == 0 || receivedApiKeys[0] != apiKey {
			return nil, status.Errorf(codes.Unauthenticated, "invalid API key")
		}

		// Continue handling the request
		return handler(ctx, req)
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
		inputs[i], err = typeconverters.ConvertStringToGivenType(input.Value, functionDefinition.Input[i].GoType)
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
		value, err := typeconverters.ConvertGivenTypeToString(result.Interface(), functionDefinition.Output[i].GoType)
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
		inputs[i], err = typeconverters.ConvertStringToGivenType(input.Value, functionDefinition.Input[i].GoType)
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
	var previousOutput *allieflowkitgrpc.StreamOutput
	for message := range *streamChannel {
		// create output
		output := &allieflowkitgrpc.StreamOutput{
			MessageCounter: counter,
			IsLast:         false,
			Value:          message,
		}

		// send output to stream
		if counter > 0 {
			err := stream.Send(previousOutput)
			if err != nil {
				return err
			}
		}

		// save output to previous output
		previousOutput = output

		// increment counter
		counter++
	}

	// send last message
	output := &allieflowkitgrpc.StreamOutput{
		MessageCounter: counter,
		IsLast:         true,
		Value:          previousOutput.Value,
	}
	err = stream.Send(output)
	if err != nil {
		return err
	}

	return nil
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
