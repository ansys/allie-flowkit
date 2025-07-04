.. _grpc:

GRPC Server & API
=================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: GRPC Service Interface
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Flowkit exposes a GRPC service for external function execution:

      - ``CallFunction`` — Executes a function and returns a single response.
      - ``CallFunctionStream`` — Executes a stream-capable function and returns a streaming response.
      - ``ListFunctions`` — Returns metadata about all registered functions.

      These methods are available to the AALI Agent and any GRPC-compatible client.

   .. grid-item-card:: Proto Contract
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      The GRPC service and message formats are defined in the ``proto/externalfunctions.proto`` file, shared via the ``aali-sharedtypes`` repository.

      Example service definition:

      .. code-block:: proto

         service ExternalFunctionService {
           rpc CallFunction (FunctionCallRequest) returns (FunctionCallResponse);
           rpc CallFunctionStream (FunctionCallRequest) returns (stream FunctionCallResponse);
           rpc ListFunctions (Empty) returns (FunctionList);
         }

      Request and response messages include fields for the function name, input arguments, outputs, and errors.

      These proto definitions are shared by both Flowkit and the AALI Agent.
