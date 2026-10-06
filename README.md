# Notification Producer

Notification Producer is a Go-based microservice that receives notification requests through REST APIs and publishes them to Azure Service Bus for asynchronous processing.

## Architecture

```text
Client
   |
   v
Notification Producer
   |
   v
Azure Service Bus Queue
   |
   v
Notification Consumer
   |
   +-- Worker Pool
   |
   v
PostgreSQL
