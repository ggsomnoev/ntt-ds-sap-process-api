```plantuml
@startuml
actor User

rectangle ProcessAPI #line:darkblue {  
  rectangle "Producer CMD" {
    component "WebAPI" as ProducerWebAPI
    component "ProcessLoader" as ProcessLoader
  }

  rectangle "Consumer CMD" {
    component "WebAPI" as ConsumerWebAPI
    component "ProcessConsumer" as ProcessConsumer
  }

  database "ProcessDB" as ProcessDB
  queue "RabbitMQ" as CommandQueue
}

User --> ProducerWebAPI : For sending process requests
User --> ConsumerWebAPI : Queries and controls processes

ProducerWebAPI --> CommandQueue : Publishes processes
ProcessLoader --> ProducerWebAPI : Updates process config information
ProcessConsumer --> CommandQueue : Consumes processes
ProcessConsumer --> ProcessDB : Writes logs, statuses
ConsumerWebAPI --> ProcessDB : Reads logs, statuses

@enduml
```