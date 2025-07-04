.. _flowkit_setup:

Setup
=====

This section guides you through installing, building, and running the Flowkit GRPC server, and connecting to it from the AALI Agent.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Prerequisites
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      - **Go 1.20+** installed on your system.
      - `git` for cloning repositories.
      - (Optional but recommended) A working `GOPATH` and Go module support enabled.

   .. grid-item-card:: Install Flowkit
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Clone the Flowkit repository and navigate to the directory:

      .. code-block:: bash

         git clone https://github.com/your-org/aali-flowkit.git
         cd aali-flowkit

      Download Go dependencies (if needed):

      .. code-block:: bash

         go mod tidy

   .. grid-item-card:: Build the Server
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Build the Go binary:

      .. code-block:: bash

         go build -o flowkit main.go

      If you encounter build errors, ensure you have the correct Go version and run `go mod tidy` to install dependencies.

   .. grid-item-card:: Run the Server
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Start the GRPC server by executing the built binary:

      .. code-block:: bash

         ./flowkit

      By default, the server listens on port **50051**.

      **Verifying the Server:**
      On successful startup, you should see a message similar to:

      .. code-block:: text

         Flowkit server started on port 50051

      If you encounter errors, check that the port is available and dependencies are up to date.

   .. grid-item-card:: Agent Connection
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Once running, Flowkit accepts GRPC requests from the AALI Agent or any GRPC-compatible client.

      - Ensure the function name and input parameters match expected definitions in your workflow.
      - Available methods are defined in ``pkg/externalfunctions/externalfunctions.go``.
      - The server responds with a result or a streamed set of messages.

   .. grid-item-card:: Next Steps
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left
