.. _execution_context:

Execution Context
=================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Context Object
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Every function in Flowkit receives a `context.Context` object when executed. This object provides:

      - Cancellation propagation
      - Deadlines and timeouts
      - Request-scoped values

      Use it to enforce execution limits or access scoped metadata.

   .. grid-item-card:: Input / Output Handling
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Functions receive input in the form of a generic `map[string]interface{}`.

      - Input data is passed as a JSON-compatible object
      - Output is returned using the same format
      - Intermediate state can be stored in logs or custom fields

   .. grid-item-card:: Shared State (Optional)
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Although Flowkit doesn't persist memory across calls by default, function authors can implement stateful logic using:

      - In-memory stores (for dev/debug)
      - External databases or cache systems
      - Streamed outputs for partial state updates

      Best practice: treat each function call as **idempotent** and stateless unless necessary.
