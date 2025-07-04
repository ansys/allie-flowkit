.. _flowkit_server:

Flowkit GRPC Server
===================

This guide explains how to run the Flowkit GRPC server (locally or with Docker), verify that it's operational, and understand its integration pattern for handling external function requests.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Run Server Locally
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      After building the binary, start the GRPC server with:

      .. code-block:: bash

         ./flowkit

      By default, the server listens on port **50051**.

      If you see an error, check that no other process is using the port and that the build succeeded.

   .. grid-item-card:: Run with Docker
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      If your repository includes a Dockerfile, you can run Flowkit with:

      .. code-block:: bash

         docker build -t aali-flowkit .
         docker run -p 50051:50051 aali-flowkit

      This will start the GRPC service at `localhost:50051`.

      **Prerequisite:**
      Ensure Docker is installed and running on your system.

   .. grid-item-card:: Test GRPC Connection (grpcurl)
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      To verify that the server is running, use [`grpcurl`](https://github.com/fullstorydev/grpcurl`):

      1. **Install grpcurl** (if not already installed):

         .. code-block:: bash

            go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

         Or use Homebrew on macOS:

         .. code-block:: bash

            brew install grpcurl

      2. **List available GRPC services:**

         .. code-block:: bash

            grpcurl -plaintext localhost:50051 list

         **Expected Output:**

         .. code-block:: text

            externalfunctions.ExternalFunctionService
            grpc.reflection.v1alpha.ServerReflection
            ...

         If you do not see these, ensure the server is running and listening on the correct port.

      3. **Troubleshooting:**
         - If you see a connection error, check that Flowkit is running and the port is not blocked.
         - If the method list is empty, ensure you have registered your functions correctly in ``pkg/externalfunctions/externalfunctions.go``.

   .. grid-item-card:: Integration Pattern
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Flowkit operates as a standalone GRPC service and typically connects to:

      - The AALI Agent, which dispatches function requests
      - External tools or developers via GRPC clients (such as `grpcurl`)

      When running, any registered function in ``externalfunctions.go`` becomes callable over GRPC.
