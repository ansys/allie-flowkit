.. _calling_functions:

Calling Registered Functions
============================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Function Invocation
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Once registered, functions in Flowkit can be called via the GRPC API. There are two primary methods:

      - `RunFunction`: Executes a function and returns a single response.
      - `StreamFunction`: Executes a function and returns a stream of responses.

      These modes support both synchronous and real-time use cases.

      .. code-block:: bash

         grpcurl -plaintext -d '{"function": "add", "input": {...}}' \
           localhost:50051 flowkit.FlowkitService/RunFunction

         grpcurl -plaintext -d '{"function": "stream_data"}' \
           localhost:50051 flowkit.FlowkitService/StreamFunction

   .. grid-item-card:: Request and Response Format
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Each request includes:

      - `function`: Name of the registered function
      - `input`: Key-value input map
      - `session_id`: (optional) Session identifier
      - `user`: (optional) User ID

      Responses include:

      - `output`: Result or stream output
      - `logs`: Structured log messages (optional)
