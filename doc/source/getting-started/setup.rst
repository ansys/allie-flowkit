.. _flowkit_setup:

Setup
=====

This section explains how to install, build, and run the Flowkit GRPC server, and how to connect to it from the AALI Agent.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Install Flowkit
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Clone the Flowkit repository and navigate to the directory:

      .. code-block:: bash

         git clone https://github.com/your-org/aali-flowkit.git
         cd aali-flowkit

      Then build the Go binary:

      .. code-block:: bash

         go build -o flowkit main.go

   .. grid-item-card:: Run the Server
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Start the GRPC server by executing the built binary:

      .. code-block:: bash

         ./flowkit

      By default, the server listens on port `50051`.

   .. grid-item-card:: Agent Connection
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Once running, Flowkit accepts GRPC requests from the AALI Agent or any other GRPC-compatible client.

      - Ensure the function name and input parameters match expected definitions
      - The server will respond with a result or a streamed set of messages
