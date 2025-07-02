.. _api_reference:

API Reference
=============

.. note::
   This page is a placeholder for auto-generated API documentation in development.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Public Interfaces
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Flowkit exposes the following core packages:

      - `pkg/function` — function registry and execution logic
      - `pkg/external_function_system/grpcserver` — GRPC server lifecycle and handlers
      - `proto/` — GRPC service definitions and contracts

   .. grid-item-card:: RegisterFunction API
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      .. code-block:: go

         func RegisterFunction(name string, fn *Function) error

      Used to attach a Go function to the Flowkit runtime registry.

   .. grid-item-card:: Function Signature
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      .. code-block:: go

         type Function struct {
           Name        string
           Description string
           Run         func(context.Context, *Request) (*Response, error)
           Stream      bool
           RunStream   func(context.Context, *Request, ServerStream) error
         }

      All functions must implement the `Run` handler.
      If `Stream` is `true`, then `RunStream` should also be defined.
