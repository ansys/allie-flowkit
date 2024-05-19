// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v5.27.0--rc3
// source: pkg/protos/externalfunctions.proto

package grpcdefinition

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ListFunctionsRequest is the input message for the ListFunctions method.
// As no input is required, this message is empty.
type ListFunctionsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListFunctionsRequest) Reset() {
	*x = ListFunctionsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFunctionsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFunctionsRequest) ProtoMessage() {}

func (x *ListFunctionsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFunctionsRequest.ProtoReflect.Descriptor instead.
func (*ListFunctionsRequest) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{0}
}

// ListFunctionsResponse is the output message for the ListFunctions method.
// It contains a map of function names to their definitions.
type ListFunctionsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Map of function names to their definitions.
	Functions map[string]*FunctionDefinition `protobuf:"bytes,1,rep,name=functions,proto3" json:"functions,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ListFunctionsResponse) Reset() {
	*x = ListFunctionsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFunctionsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFunctionsResponse) ProtoMessage() {}

func (x *ListFunctionsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFunctionsResponse.ProtoReflect.Descriptor instead.
func (*ListFunctionsResponse) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{1}
}

func (x *ListFunctionsResponse) GetFunctions() map[string]*FunctionDefinition {
	if x != nil {
		return x.Functions
	}
	return nil
}

// FunctionDefinition is the definition of an individual function.
// It contains the name, description, inputs and outputs of the function.
type FunctionDefinition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the function.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Description of the function.
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// List of input definitions for the function.
	Input []*FunctionInputDefinition `protobuf:"bytes,3,rep,name=input,proto3" json:"input,omitempty"`
	// List of output definitions for the function.
	Output []*FunctionOutputDefinition `protobuf:"bytes,4,rep,name=output,proto3" json:"output,omitempty"`
}

func (x *FunctionDefinition) Reset() {
	*x = FunctionDefinition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionDefinition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionDefinition) ProtoMessage() {}

func (x *FunctionDefinition) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionDefinition.ProtoReflect.Descriptor instead.
func (*FunctionDefinition) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{2}
}

func (x *FunctionDefinition) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FunctionDefinition) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *FunctionDefinition) GetInput() []*FunctionInputDefinition {
	if x != nil {
		return x.Input
	}
	return nil
}

func (x *FunctionDefinition) GetOutput() []*FunctionOutputDefinition {
	if x != nil {
		return x.Output
	}
	return nil
}

// FunctionInputDefinition is the definition of an input for a function.
// It contains the name, type, Go language type and options for the input.
type FunctionInputDefinition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the input.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Type of the input.
	Type string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	// Go language type of the input.
	GoType string `protobuf:"bytes,3,opt,name=go_type,json=goType,proto3" json:"go_type,omitempty"`
	// List of options for the input, if applicable.
	Options []string `protobuf:"bytes,4,rep,name=options,proto3" json:"options,omitempty"`
}

func (x *FunctionInputDefinition) Reset() {
	*x = FunctionInputDefinition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionInputDefinition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionInputDefinition) ProtoMessage() {}

func (x *FunctionInputDefinition) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionInputDefinition.ProtoReflect.Descriptor instead.
func (*FunctionInputDefinition) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{3}
}

func (x *FunctionInputDefinition) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FunctionInputDefinition) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *FunctionInputDefinition) GetGoType() string {
	if x != nil {
		return x.GoType
	}
	return ""
}

func (x *FunctionInputDefinition) GetOptions() []string {
	if x != nil {
		return x.Options
	}
	return nil
}

// FunctionOutputDefinition is the definition of an output for a function.
// It contains the name, type and Go language type of the output.
type FunctionOutputDefinition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the output.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Type of the output.
	Type string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	// Go language type of the output.
	GoType string `protobuf:"bytes,3,opt,name=go_type,json=goType,proto3" json:"go_type,omitempty"`
}

func (x *FunctionOutputDefinition) Reset() {
	*x = FunctionOutputDefinition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionOutputDefinition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionOutputDefinition) ProtoMessage() {}

func (x *FunctionOutputDefinition) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionOutputDefinition.ProtoReflect.Descriptor instead.
func (*FunctionOutputDefinition) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{4}
}

func (x *FunctionOutputDefinition) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FunctionOutputDefinition) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *FunctionOutputDefinition) GetGoType() string {
	if x != nil {
		return x.GoType
	}
	return ""
}

// FunctionInputs is the input message for the RunFunction method.
// It contains the name of the function to run and a list of inputs.
type FunctionInputs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the function to run.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// List of inputs for the function.
	Inputs []*FunctionInput `protobuf:"bytes,2,rep,name=inputs,proto3" json:"inputs,omitempty"`
}

func (x *FunctionInputs) Reset() {
	*x = FunctionInputs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionInputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionInputs) ProtoMessage() {}

func (x *FunctionInputs) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionInputs.ProtoReflect.Descriptor instead.
func (*FunctionInputs) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{5}
}

func (x *FunctionInputs) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FunctionInputs) GetInputs() []*FunctionInput {
	if x != nil {
		return x.Inputs
	}
	return nil
}

// Single input for a function.
type FunctionInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the input.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Go language type of the input.
	GoType string `protobuf:"bytes,2,opt,name=go_type,json=goType,proto3" json:"go_type,omitempty"`
	// Value of the input.
	Value string `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *FunctionInput) Reset() {
	*x = FunctionInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionInput) ProtoMessage() {}

func (x *FunctionInput) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionInput.ProtoReflect.Descriptor instead.
func (*FunctionInput) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{6}
}

func (x *FunctionInput) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FunctionInput) GetGoType() string {
	if x != nil {
		return x.GoType
	}
	return ""
}

func (x *FunctionInput) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

// FunctionOutputs is the output message for the RunFunction method.
// It contains the name of the function that was run and a list of outputs.
type FunctionOutputs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the function that was run.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// List of outputs from the function.
	Outputs []*FunctionOutput `protobuf:"bytes,2,rep,name=outputs,proto3" json:"outputs,omitempty"`
}

func (x *FunctionOutputs) Reset() {
	*x = FunctionOutputs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionOutputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionOutputs) ProtoMessage() {}

func (x *FunctionOutputs) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionOutputs.ProtoReflect.Descriptor instead.
func (*FunctionOutputs) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{7}
}

func (x *FunctionOutputs) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FunctionOutputs) GetOutputs() []*FunctionOutput {
	if x != nil {
		return x.Outputs
	}
	return nil
}

// FunctionOutput is a single output from a function.
// It contains the name, Go language type and value of the output.
type FunctionOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the output.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Go language type of the output.
	GoType string `protobuf:"bytes,2,opt,name=go_type,json=goType,proto3" json:"go_type,omitempty"`
	// Value of the output.
	Value string `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *FunctionOutput) Reset() {
	*x = FunctionOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionOutput) ProtoMessage() {}

func (x *FunctionOutput) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionOutput.ProtoReflect.Descriptor instead.
func (*FunctionOutput) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{8}
}

func (x *FunctionOutput) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FunctionOutput) GetGoType() string {
	if x != nil {
		return x.GoType
	}
	return ""
}

func (x *FunctionOutput) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

// StreamOutput is the output message for the StreamFunction method.
// It contains the message counter, a flag indicating if this is the last message
type StreamOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Counter for the message in the stream.
	MessageCounter int32 `protobuf:"varint,1,opt,name=message_counter,json=messageCounter,proto3" json:"message_counter,omitempty"`
	// Indicates if this is the last message in the stream.
	IsLast bool `protobuf:"varint,2,opt,name=is_last,json=isLast,proto3" json:"is_last,omitempty"`
	// Value of the output.
	Value string `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *StreamOutput) Reset() {
	*x = StreamOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_protos_externalfunctions_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamOutput) ProtoMessage() {}

func (x *StreamOutput) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_protos_externalfunctions_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamOutput.ProtoReflect.Descriptor instead.
func (*StreamOutput) Descriptor() ([]byte, []int) {
	return file_pkg_protos_externalfunctions_proto_rawDescGZIP(), []int{9}
}

func (x *StreamOutput) GetMessageCounter() int32 {
	if x != nil {
		return x.MessageCounter
	}
	return 0
}

func (x *StreamOutput) GetIsLast() bool {
	if x != nil {
		return x.IsLast
	}
	return false
}

func (x *StreamOutput) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_pkg_protos_externalfunctions_proto protoreflect.FileDescriptor

var file_pkg_protos_externalfunctions_proto_rawDesc = []byte{
	0x0a, 0x22, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x65, 0x78, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x16, 0x0a, 0x14, 0x4c, 0x69, 0x73, 0x74, 0x46,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22,
	0xd3, 0x01, 0x0a, 0x15, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x55, 0x0a, 0x09, 0x66, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x37, 0x2e, 0x65,
	0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x1a, 0x63, 0x0a, 0x0e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x3b, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xd1, 0x01, 0x0a, 0x12, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x40, 0x0a, 0x05, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x2a, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e,
	0x70, 0x75, 0x74, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x69,
	0x6e, 0x70, 0x75, 0x74, 0x12, 0x43, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x22, 0x74, 0x0a, 0x17, 0x46, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x17, 0x0a, 0x07,
	0x67, 0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x67,
	0x6f, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22,
	0x5b, 0x0a, 0x18, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x67, 0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x67, 0x6f, 0x54, 0x79, 0x70, 0x65, 0x22, 0x5e, 0x0a, 0x0e,
	0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x38, 0x0a, 0x06, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x20, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x52, 0x06, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x22, 0x52, 0x0a, 0x0d,
	0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x17, 0x0a, 0x07, 0x67, 0x6f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x67, 0x6f, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x22, 0x62, 0x0a, 0x0f, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x3b, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x46, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x07, 0x6f, 0x75, 0x74,
	0x70, 0x75, 0x74, 0x73, 0x22, 0x53, 0x0a, 0x0e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x67, 0x6f,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x67, 0x6f, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x66, 0x0a, 0x0c, 0x53, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74,
	0x65, 0x72, 0x12, 0x17, 0x0a, 0x07, 0x69, 0x73, 0x5f, 0x6c, 0x61, 0x73, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x06, 0x69, 0x73, 0x4c, 0x61, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x32, 0xab, 0x02, 0x0a, 0x11, 0x45, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x46, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x64, 0x0a, 0x0d, 0x4c, 0x69, 0x73, 0x74, 0x46,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x27, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x28, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x56, 0x0a,
	0x0b, 0x52, 0x75, 0x6e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x21, 0x2e, 0x65,
	0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x1a,
	0x22, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x73, 0x22, 0x00, 0x12, 0x58, 0x0a, 0x0e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x46,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x21, 0x2e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x46, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x1a, 0x1f, 0x2e, 0x65, 0x78, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x22, 0x00, 0x30, 0x01, 0x42,
	0x12, 0x5a, 0x10, 0x2e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_protos_externalfunctions_proto_rawDescOnce sync.Once
	file_pkg_protos_externalfunctions_proto_rawDescData = file_pkg_protos_externalfunctions_proto_rawDesc
)

func file_pkg_protos_externalfunctions_proto_rawDescGZIP() []byte {
	file_pkg_protos_externalfunctions_proto_rawDescOnce.Do(func() {
		file_pkg_protos_externalfunctions_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_protos_externalfunctions_proto_rawDescData)
	})
	return file_pkg_protos_externalfunctions_proto_rawDescData
}

var file_pkg_protos_externalfunctions_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_pkg_protos_externalfunctions_proto_goTypes = []interface{}{
	(*ListFunctionsRequest)(nil),     // 0: externalfunctions.ListFunctionsRequest
	(*ListFunctionsResponse)(nil),    // 1: externalfunctions.ListFunctionsResponse
	(*FunctionDefinition)(nil),       // 2: externalfunctions.FunctionDefinition
	(*FunctionInputDefinition)(nil),  // 3: externalfunctions.FunctionInputDefinition
	(*FunctionOutputDefinition)(nil), // 4: externalfunctions.FunctionOutputDefinition
	(*FunctionInputs)(nil),           // 5: externalfunctions.FunctionInputs
	(*FunctionInput)(nil),            // 6: externalfunctions.FunctionInput
	(*FunctionOutputs)(nil),          // 7: externalfunctions.FunctionOutputs
	(*FunctionOutput)(nil),           // 8: externalfunctions.FunctionOutput
	(*StreamOutput)(nil),             // 9: externalfunctions.StreamOutput
	nil,                              // 10: externalfunctions.ListFunctionsResponse.FunctionsEntry
}
var file_pkg_protos_externalfunctions_proto_depIdxs = []int32{
	10, // 0: externalfunctions.ListFunctionsResponse.functions:type_name -> externalfunctions.ListFunctionsResponse.FunctionsEntry
	3,  // 1: externalfunctions.FunctionDefinition.input:type_name -> externalfunctions.FunctionInputDefinition
	4,  // 2: externalfunctions.FunctionDefinition.output:type_name -> externalfunctions.FunctionOutputDefinition
	6,  // 3: externalfunctions.FunctionInputs.inputs:type_name -> externalfunctions.FunctionInput
	8,  // 4: externalfunctions.FunctionOutputs.outputs:type_name -> externalfunctions.FunctionOutput
	2,  // 5: externalfunctions.ListFunctionsResponse.FunctionsEntry.value:type_name -> externalfunctions.FunctionDefinition
	0,  // 6: externalfunctions.ExternalFunctions.ListFunctions:input_type -> externalfunctions.ListFunctionsRequest
	5,  // 7: externalfunctions.ExternalFunctions.RunFunction:input_type -> externalfunctions.FunctionInputs
	5,  // 8: externalfunctions.ExternalFunctions.StreamFunction:input_type -> externalfunctions.FunctionInputs
	1,  // 9: externalfunctions.ExternalFunctions.ListFunctions:output_type -> externalfunctions.ListFunctionsResponse
	7,  // 10: externalfunctions.ExternalFunctions.RunFunction:output_type -> externalfunctions.FunctionOutputs
	9,  // 11: externalfunctions.ExternalFunctions.StreamFunction:output_type -> externalfunctions.StreamOutput
	9,  // [9:12] is the sub-list for method output_type
	6,  // [6:9] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_pkg_protos_externalfunctions_proto_init() }
func file_pkg_protos_externalfunctions_proto_init() {
	if File_pkg_protos_externalfunctions_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_protos_externalfunctions_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFunctionsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFunctionsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionDefinition); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionInputDefinition); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionOutputDefinition); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionInputs); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionInput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionOutputs); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_protos_externalfunctions_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_protos_externalfunctions_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_protos_externalfunctions_proto_goTypes,
		DependencyIndexes: file_pkg_protos_externalfunctions_proto_depIdxs,
		MessageInfos:      file_pkg_protos_externalfunctions_proto_msgTypes,
	}.Build()
	File_pkg_protos_externalfunctions_proto = out.File
	file_pkg_protos_externalfunctions_proto_rawDesc = nil
	file_pkg_protos_externalfunctions_proto_goTypes = nil
	file_pkg_protos_externalfunctions_proto_depIdxs = nil
}
