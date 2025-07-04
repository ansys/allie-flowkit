.. _exposed_variables:

Exposed Variables
=================

This section explains how to share variables across workflow steps in Flowkit using the GRPC `input` and `output` maps, with examples of scoping and usage in Go.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Variable Sharing Across Steps
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Workflows often need to reuse values across nodes.
      Flowkit supports passing key-value pairs between steps using the GRPC ``input`` and ``output`` maps.

      **Example:**
      - Function A extracts a value and outputs ``{"customer_id": "123"}``
      - Function B receives this value as part of its input in the same session

      Variables are available to all functions within a single workflow session.

   .. grid-item-card:: Key Scoping & Reserved Keys
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Variables are scoped **per session**.
      The same key can exist in different sessions without conflict.

      **Reserved keys** (automatically set by the AALI Agent):
      - ``__user``: The current user identifier
      - ``__session``: The session ID for the workflow execution
      - ``__workflow``: The workflow ID

      These keys are injected into each function request for tracking and audit purposes.

      **Note:**
      Avoid overwriting reserved keys, as they may be required for session continuity and security.

   .. grid-item-card:: Access Pattern in Go
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Access input variables using Go map syntax in your function implementation.

      .. code-block:: go

         // req is *flowkitpb.FunctionRequest
         customerID, ok := req.Input["customer_id"].GetStringValue()
         if !ok {
             return nil, errors.New("missing customer_id")
         }

      Set output variables using the Flowkit response format:

      .. code-block:: go

         // Returning output variables
         return &flowkitpb.FunctionResponse{
             Output: map[string]*structpb.Value{
                 "status": structpb.NewStringValue("ok"),
                 "next":   structpb.NewStringValue("complete"),
             },
         }

      For full proto definitions, see ``proto/externalfunctions.proto``.

   .. grid-item-card:: Example: Multi-Step Workflow
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      **Step 1: Extract a variable**

      .. code-block:: go

         // FunctionA outputs a value
         return &flowkitpb.FunctionResponse{
             Output: map[string]*structpb.Value{
                 "customer_id": structpb.NewStringValue("123"),
             },
         }

      **Step 2: Consume the variable**

      .. code-block:: go

         // FunctionB receives the value as input in the same session
         customerID, ok := req.Input["customer_id"].GetStringValue()
         // Use customerID for further processing

      All variable passing is handled automatically by the workflow session context.
