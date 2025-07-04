.. _agent_integration:

Agent Integration
=================

This section explains how the AALI Agent communicates with Flowkit over GRPC and provides practical guidance for integration and debugging.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Communication Overview
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      The AALI Agent connects to Flowkit via GRPC to execute workflow steps. For each step, the Agent can invoke one or more registered functions.

      **Call sequence:**

      1. Agent selects a function to invoke
      2. Sends a GRPC request with input arguments
      3. Flowkit executes the function and returns a response
      4. Agent processes the result and continues the workflow

      Both synchronous and streaming (multi-response) functions are supported.

   .. grid-item-card:: Proto Definitions & Available Methods
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      All available methods are defined in ``pkg/externalfunctions/externalfunctions.go``.

      The GRPC service interface is defined in the ``proto/externalfunctions.proto`` file.

      Example function call (from the Agent's perspective):

      .. code-block:: text

         service ExternalFunctionService {
             rpc CallFunction (FunctionCallRequest) returns (FunctionCallResponse);
         }

      See the proto file for request/response message details.

   .. grid-item-card:: Agent Configuration Example
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      To connect the Agent to a running Flowkit instance, set the Flowkit GRPC endpoint in the Agent’s configuration file or environment variables.

      Example config (YAML):

      .. code-block:: yaml

         flowkit:
           address: "localhost:50051"

      Or set the environment variable:

      .. code-block:: bash

         export FLOWKIT_GRPC_ADDR=localhost:50051

   .. grid-item-card:: Direct GRPC Access & Debugging
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Any client can interact with Flowkit using its GRPC API (not just the Agent):

      - Use `grpcurl`, custom scripts, or test clients for debugging and development.
      - Example (list available services):

        .. code-block:: bash

           grpcurl -plaintext localhost:50051 list

      For troubleshooting connection errors, check that:
      - The server is running and reachable on the configured port
      - The proto file matches the running Flowkit server version
