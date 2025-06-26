.. _flowkit_setup:

Setup
=====

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Install Flowkit
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Clone the Flowkit repository and navigate to the directory:

      .. code-block:: bash

         git clone https://github.com/your-org/aali-flowkit.git
         cd aali-flowkit

      You can now build the Go binary:

      .. code-block:: bash

         go build -o flowkit main.go

   .. grid-item-card:: Run the Server
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Start the GRPC server by running the built binary:

      .. code-block:: bash

         ./flowkit

      The server will listen on port `50051` by default.

   .. grid-item-card:: Agent Connection
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Once running, Flowkit accepts GRPC requests from the AALI Agent or any other GRPC client.

      - Requests should specify the function name and input parameters
      - The response will contain the result or a stream of messages