.. _troubleshooting:

Common Issues
=============

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: GRPC Endpoint Not Responding
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      If `grpcurl` or any client fails to connect:

      - Check that Flowkit is running and listening on `:50051`
      - Confirm port forwarding if running inside Docker
      - Ensure firewall or network rules don’t block access

   .. grid-item-card:: "function not registered" Error
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      This error indicates a missing or misnamed function. Make sure:

      - You called `RegisterFunction()` in `main.go`
      - The function name exactly matches the GRPC request
      - The function is included before the GRPC server starts

   .. grid-item-card:: Empty or Broken Streaming Output
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      When using `StreamFunction`, your Go function must:

      - Have `Stream: true` in the registration metadata
      - Use a goroutine to emit multiple `FunctionResponse` messages
      - Handle `ctx` cancellation to exit early when the client disconnects

   .. grid-item-card:: Agent Doesn’t Respond
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      If Flowkit forwards calls to the Agent and nothing happens:

      - Verify the Agent is running and reachable
      - Check logs on both Agent and Flowkit for errors
      - Confirm that `input` contains all required fields for the Agent logic
