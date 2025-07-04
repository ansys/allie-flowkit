.. _calling_functions:

Calling Registered Functions
============================

This section explains how to invoke registered functions in Flowkit using the GRPC API, with examples for both synchronous and streaming calls.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Function Invocation Methods
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Registered functions can be called using the GRPC API. Two primary methods are provided by the ``ExternalFunctionService``:

      - ``CallFunction``: Executes a function and returns a single response.
      - ``CallFunctionStream``: Executes a function and returns a stream of responses.

      These modes support both synchronous and real-time scenarios.

      Example usage with [`grpcurl`](https://github.com/fullstorydev/grpcurl):

      .. code-block:: bash

         grpcurl -plaintext -d '{
           "function": "TransformData",
           "input": {"dataform": "sample input", "depth": 2}
         }' \
         localhost:50051 externalfunctions.ExternalFunctionService/CallFunction

         grpcurl -plaintext -d '{
           "function": "StreamData",
           "input": {"param": "value"}
         }' \
         localhost:50051 externalfunctions.ExternalFunctionService/CallFunctionStream

   .. grid-item-card:: Request and Response Format
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      **Request (FunctionCallRequest):**

      - ``function``: Name of the registered function (string)
      - ``input``: Key-value map of function arguments (object)
      - ``session_id``: (optional) Session identifier (string)
      - ``user``: (optional) User ID (string)

      **Response (FunctionCallResponse):**

      - ``output``: Function result or stream output
      - ``logs``: Structured log messages (optional)
      - ``error``: Error message if function fails

      See ``proto/externalfunctions.proto`` for exact message structure.

      **Note:**
      For a list of available functions, see ``pkg/externalfunctions/externalfunctions.go``.
      If you call a function that is not registered, you will receive an error in the response.

      Example error response:

      .. code-block:: json

         {
           "error": "function not found: MyMissingFunction"
         }
