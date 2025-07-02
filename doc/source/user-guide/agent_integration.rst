.. _agent_integration:

Agent Integration
=================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Communication with Agent
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      The AALI Agent connects to Flowkit via GRPC. For each workflow step, the Agent can invoke one or more registered functions.

      Call sequence:

      1. Agent selects a function to invoke
      2. Sends a GRPC request with arguments
      3. Flowkit executes the function
      4. Agent processes the result and continues the workflow

      Both synchronous and streaming functions are supported.

   .. grid-item-card:: Direct GRPC Access
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Although primarily used with the AALI Agent, any client can interact with Flowkit via its GRPC API.

      - Use `grpcurl`, custom clients, or test scripts
      - Helpful for debugging or running standalone setups
