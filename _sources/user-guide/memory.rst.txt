.. _execution_context:

Execution Context
=================

This section describes the environment and conventions when writing functions for Flowkit, including the context object, parameter handling, and state management best practices.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Context Object
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Every function in Flowkit receives a `context.Context` object as the first parameter.
      This object provides:

      - Cancellation propagation (for timeouts or aborted workflows)
      - Deadlines and timeouts (set upstream)
      - Request-scoped values and metadata

      Use the context object to:
      - Enforce execution limits
      - Access request metadata (e.g., user/session) if needed

      Example usage:

      .. code-block:: go

         func MyFunction(ctx context.Context, input string) (output string, err error) {
             select {
             case <-ctx.Done():
                 return "", ctx.Err()
             default:
                 // Your logic here
             }
         }

   .. grid-item-card:: Input and Output Handling
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Functions in Flowkit define **typed arguments** (not just `map[string]interface{}`):

      .. code-block:: go

         func TransformData(dataform string, depth int) (transformed string, err error)

      When invoked over GRPC, the input arguments are mapped from the request message fields, and the outputs are returned as response fields.

      - Input and output values should be **JSON-serializable** when needed (for interoperability)
      - See `pkg/externalfunctions/externalfunctions.go` and the proto definitions for details.

      Logs and debug output can be returned as part of the response if desired.

   .. grid-item-card:: State and Session Context
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      By default, each function call is **stateless and idempotent**.
      Flowkit does not persist in-memory state across calls.

      To manage state or share data across workflow steps:
      - Use input/output variables via the GRPC maps (see :ref:`Exposed Variables <exposed_variables>`)
      - For more advanced use cases, connect to external databases, caches, or file systems.

      **Best Practice:**
      Treat each function as independent and stateless unless explicit session context is required.

   .. grid-item-card:: More Information
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      - See `proto/externalfunctions.proto` for message details
      - For input/output conventions, review sample functions in `pkg/externalfunctions/`
      - For workflow session context, see :ref:`Exposed Variables <exposed_variables>`
