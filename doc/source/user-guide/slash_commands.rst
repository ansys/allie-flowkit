.. _slash_commands:

Slash Commands
==============

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Slash Command Handling
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Flowkit itself does **not** handle raw slash commands like `/reset` or `/summary`.

      These commands are parsed and interpreted by the **AALI Agent** or client application.
      The agent then converts them into valid GRPC function calls and sends them to Flowkit.

      **Where to look:**
      - The logic for parsing slash commands and mapping them to function calls lives in the Agent (see the Agent repository for implementation details).

   .. grid-item-card:: Function Dispatching
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Flowkit receives normal GRPC requests where:

      - ``function`` contains the name of a registered function (e.g. `reset_session`)
      - ``input`` includes any parameters from the slash command
      - No special syntax like `/` is interpreted at the Flowkit level

      **Example:**
      The slash command `/reset` becomes a GRPC call:

      .. code-block:: json

         {
           "function": "reset_session",
           "input": {}
         }

   .. grid-item-card:: Usage Context
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Slash commands are often used in:

      - Developer tooling or chat interfaces
      - AALI Agent workflows for fast operator actions
      - Automating tasks like session resets or translations

      Once invoked, Flowkit executes the registered function as a standard GRPC request.
