.. _function_registration:

Function Registration
=====================

This section explains how to add and register new functions in Flowkit so they are available for external calls via the GRPC API.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Adding a New Function
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      To expose a new function in Flowkit:

      1. **Define your function** in the appropriate package inside `pkg/externalfunctions/`.
      2. **Include a `displayName` tag** in the comments for UI integration.

      Example:

      .. code-block:: go

         // File: aali-flowkit/pkg/externalfunctions/data_extraction.go

         // TransformData processes input data and returns a transformed result.
         // Tags:
         //   - @displayName: Transform the Data
         func TransformData(dataform string, depth int) (transformed string, err error) {
             // Implementation
             return "transformed_data", nil
         }

   .. grid-item-card:: Register the Function in the Registry
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Register the function in `pkg/externalfunctions/externalfunctions.go` by adding it to the `ExternalFunctionsMap`:

      .. code-block:: go

         var ExternalFunctionsMap = map[string]interface{}{
             // Existing functions...
             "TransformData": TransformData, // Add your function here
         }

      The key is the function name (as called from the agent), and the value is the Go function.

      **Note:**
      All functions must be registered here to be available via GRPC.

   .. grid-item-card:: Function Signature Requirements
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      All registered functions must use standard Go signatures (see examples in `externalfunctions/`).

      For advanced features like streaming responses, consult the proto definition and core implementation. Most functions are single-response by default.

      For more details, see the main README section:
      **Adding New Functions, Types, and Categories**
