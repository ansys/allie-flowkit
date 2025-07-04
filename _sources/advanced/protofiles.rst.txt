.. _protofiles:

Proto Definitions
=================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Shared Contracts
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      All GRPC services in AALI — including Flowkit and Agent — rely on shared `.proto` files.
      These define:

      - Function call request/response format
      - Stream interfaces
      - Service method signatures

   .. grid-item-card:: Example Snippet
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Example from ``proto/externalfunctions.proto``:

      .. code-block:: proto

         service ExternalFunctionService {
           rpc CallFunction (FunctionCallRequest) returns (FunctionCallResponse);
           rpc CallFunctionStream (FunctionCallRequest) returns (stream FunctionCallResponse);
           rpc ListFunctions (Empty) returns (FunctionList);
         }

   .. grid-item-card:: Location & Usage
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      The `.proto` files are located in the ``proto/`` directory of the `aali-sharedtypes` repository.

      To regenerate Go types:

      .. code-block:: bash

         protoc --go_out=. --go-grpc_out=. proto/externalfunctions.proto

      This ensures consistent typing for services written in Go, Python, or other supported languages.
