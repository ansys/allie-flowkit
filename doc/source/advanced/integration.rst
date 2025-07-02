.. _integration:


Agent Integration Details
=========================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: What the Agent Does
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      The AALI Agent is a backend component that interfaces with Flowkit to run workflows, maintain state, and chain outputs to downstream steps.

      It implements the GRPC contract defined in `proto/`.

   .. grid-item-card:: Execution Flow
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      1. A workflow node calls Flowkit
      2. Flowkit forwards input to the Agent
      3. Agent processes logic (via plugins or direct code)
      4. Output is streamed or returned

      Flowkit is stateless — the Agent may implement memory.

   .. grid-item-card:: Extending Agents
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Agents can support:

      - Custom slash commands
      - Persistent memory backends
      - Logging, permissions, and retry policies

      Flowkit sends function name and input map — it is up to the Agent how to process.
