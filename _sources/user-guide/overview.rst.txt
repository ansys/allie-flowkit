.. _overview:

Overview
========

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Flowkit Overview
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Flowkit is the GRPC-based execution engine behind the AALI system. It allows services to call external functions over the network using a simple GRPC interface. Each function is registered at startup and can be invoked remotely with or without stream support.

      **Responsibilities:**

      - Run as a standalone GRPC server
      - Expose registered external functions for agent access
      - Handle session context and runtime tracking
      - Power execution flow in agent conversations

      Built in Go, Flowkit is modular, fast, and easy to extend. New functionality can be plugged in via simple function registration.