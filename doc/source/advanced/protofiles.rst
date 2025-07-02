.. _protofiles:

Proto Definitions
=================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Shared Contracts
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      All GRPC services in AALI — including Flowkit and Agent — rely on shared `.proto` files. These define:

      - Function call request/response format
      - Stream interfaces
      - Service method signatures

   .. grid-item-card:: Example Snippet
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      .. code-block:: proto

         service FlowkitService {
           rpc RunFunction(FunctionRequest) returns (FunctionResponse);
           rpc StreamFunction(FunctionRequest) returns (stream FunctionResponse);
           rpc ListFunctions(Empty) returns (FunctionList);
         }

   .. grid-item-card:: Location
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      The `.proto` files live under the `proto/` directory and are compiled via `protoc`. They ensure consistent typing between services implemented in Go or Python.