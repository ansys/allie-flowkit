.. _function_registration:

Function Registration
=====================

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Registering Functions in Flowkit
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      In Flowkit, functions are modular units of logic that must be registered before they can be called over GRPC. Each function is stored in a shared registry with a unique name and optional metadata.

      **How it works:**

      - Functions are registered at startup using `flowkit.RegisterFunction(...)`
      - Each function has a name, description, and implementation logic
      - Optionally, a function can support streaming results

      The function must follow this signature:

      .. code-block:: go

         func(ctx context.Context, req *Request) (*Response, error)

      Example registration:

      .. code-block:: go

         flowkit.RegisterFunction("add", &Function{
             Name: "add",
             Description: "Adds two integers",
             Run: func(ctx context.Context, req *Request) (*Response, error) {
                 // Implementation here...
             },
         })

      To register a streaming function:

      .. code-block:: go

         flowkit.RegisterFunction("streamLogs", &Function{
             Name:   "streamLogs",
             Stream: true,
             Description: "Streams log output line by line",
             RunStream: func(ctx context.Context, req *Request, stream ServerStream) error {
                 // Stream logic...
             },
         })
