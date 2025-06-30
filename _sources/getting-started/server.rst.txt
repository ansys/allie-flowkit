.. _flowkit_server:

Flowkit GRPC Server
===========================

This guide explains how to run the Flowkit GRPC server and verify that it's operational for handling external function requests.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Run Server (Local)
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      After building the binary, start the GRPC server with:

      .. code-block:: bash

         ./flowkit

      By default, the server listens on port `50051`.

   .. grid-item-card:: Run with Docker (If Configured)
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      If a Dockerfile is configured, you can run Flowkit using:

      .. code-block:: bash

         docker build -t aali-flowkit .
         docker run -p 50051:50051 aali-flowkit

      The GRPC service will be available at `localhost:50051`.

   .. grid-item-card:: Test GRPC Connection (grpcurl)
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Use [`grpcurl`](https://github.com/fullstorydev/grpcurl) to verify the server is running:

      .. code-block:: bash

         grpcurl -plaintext localhost:50051 list

      You should see a list of available GRPC methods.

   .. grid-item-card:: Integration Pattern
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Flowkit operates as a standalone GRPC service and typically connects to:

      - The AALI Agent, which dispatches function requests
      - External tools or developers via GRPC clients (e.g. `grpcurl`)

      When running, any registered function becomes callable over GRPC.