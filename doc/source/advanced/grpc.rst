.. _grpc:

GRPC Server & API
=================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: GRPC Service Interface
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Flowkit exposes a GRPC service with three core methods:

      - `ListFunctions` — Returns metadata about all registered functions
      - `RunFunction` — Executes a standard function and returns the result
      - `StreamFunction` — Executes a stream-capable function and returns a streaming response

      These methods are used internally by the Agent, but can also be accessed by other tools.

   .. grid-item-card:: Proto Contract
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      The GRPC service is defined in shared `.proto` files in the `aali-sharedtypes` repo.

      Example:

      .. code-block:: proto

         service Flowkit {
           rpc RunFunction(FunctionRequest) returns (FunctionResponse);
           rpc StreamFunction(FunctionRequest) returns (stream FunctionResponse);
           rpc ListFunctions(Empty) returns (FunctionList);
         }

      These proto definitions are shared across Flowkit and Agent.