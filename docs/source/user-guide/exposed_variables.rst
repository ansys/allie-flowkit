.. _exposed_variables:

Exposed Variables
=================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Variable Sharing Across Steps
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Workflows often need to reuse values across nodes.  
      Flowkit supports passing key-value pairs between steps using GRPC `input` and `output` maps.

      Examples:

      - Extracting a value in Function A
      - Passing it into Function B as part of the same session

   .. grid-item-card:: Key Scoping
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Variables are scoped per session. The same keys can exist in different sessions without conflict.

      Reserved keys include:

      - `__user`
      - `__session`
      - `__workflow`

      These are injected by the AALI Agent during request initialization.

   .. grid-item-card:: Access Pattern (Go)
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Access input variables using standard Go map syntax:

      .. code-block:: go

         customerID, ok := req.Input["customer_id"].GetStringValue()
         if !ok {
             return nil, errors.New("missing customer_id")
         }

      Create output using the Flowkit response format:

      .. code-block:: go

         return &flowkitpb.FunctionResponse{
             Output: map[string]*structpb.Value{
                 "status": structpb.NewStringValue("ok"),
                 "next":   structpb.NewStringValue("complete"),
             },
         }