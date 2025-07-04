.. _troubleshooting:

Common Issues
=============

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: GRPC Endpoint Not Responding
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      If `grpcurl` or any client fails to connect:

      - Ensure Flowkit is running and listening on `:50051`
      - If using Docker, confirm the port is mapped with `-p 50051:50051`
      - Check firewall or network rules are not blocking access
      - Review Flowkit logs for any startup or bind errors

   .. grid-item-card:: "function not registered" Error
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      This error means the function name in your GRPC request is not registered in Flowkit.

      - Verify your function is added to `ExternalFunctionsMap` in `pkg/externalfunctions/externalfunctions.go`
      - The key in the map must exactly match the name sent in your request
      - Save and rebuild the Flowkit binary after adding new functions

      Example error response:

      .. code-block:: json

         { "error": "function not found: MyFunction" }

   .. grid-item-card:: Empty or Broken Streaming Output
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      If a streaming call returns no data or ends early:

      - Ensure your function supports streaming (see proto and implementation for correct signature)
      - Make sure the function is correctly mapped in `ExternalFunctionsMap`
      - Use logs to verify data is being generated and sent

      If streaming support is not yet implemented for your function, fallback to single-response mode.

   .. grid-item-card:: Agent Doesn’t Respond
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      If Flowkit forwards calls to the Agent and nothing happens:

      - Verify the Agent is running and its address is configured correctly
      - Check both Agent and Flowkit logs for errors
      - Ensure your GRPC `input` includes all required fields for the Agent logic
      - Confirm version compatibility between Flowkit and Agent

   .. grid-item-card:: Still Stuck?
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      - Review logs in `logs/` or console output for error messages
      - Confirm your proto definitions match between client and server
      - See the README or :ref:`Setup <flowkit_setup>` for environment troubleshooting tips
