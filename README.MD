# Process API

## Description
This app is a tiny demonstration that shows how to manage and run process commands:
* Accepts **HTTP** requests to start process commands. Each command can: 
  - Run locally
  - Run on a remote machine via SSH
  - Copy files to a remote machine using SFTP 
* Pushes every request onto a **RabbitMQ** queue.
* Retries messages on failures couple of times. 
* **Consumer/s** processes the queue and performs the required actions based on the command type (localCmd, sshCmd, scpCmd).
* Persists every processed/ing message in PostgreSQL:
  * used for tracking.
  * logs and states can be retrieved later through the API.
* Each command must:
  - Specify a **process name** that exists in the **configured process definitions**. This can be specified in the '.env' file. The `process_configs` project directory will be used by default.
  - Include all **required parameters** expected by that process

## Architecture

The project follows [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/The-Clean-Architecture.html) to get:

* **easy unit-testing & mocking**  
* **clear separation of concerns**  
* **plug-and-play transports or providers** 

Key libraries & tools:

| Concern | Library |
|---------|---------|
| HTTP server | [Echo](https://echo.labstack.com/) |
| DB migrations | [golang-migrate/migrate](https://github.com/golang-migrate/migrate) |
| Testing | [Ginkgo](https://onsi.github.io/ginkgo/) + [Gomega](https://onsi.github.io/gomega/) |
| Fakes | [Counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) |
| Queue | RabbitMQ (`amqp091-go`) |

Useful links:
* [C4 Diagram](./docs/c4diagram.md)

## How to run the server

To run the server using docker:
```bash
make run-docker
```

If you want to manually run the migrations:

```bash
make migrate-down
make migrate-up
```

## How to run the tests

To run the unit tests:

```bash
make test
```

## Example requests and responses

### To execute command locally

```bash
curl -X POST http://127.0.0.1:8080/startProcess \
  -H "Content-Type: application/json" \
  -d '{
        "name": "uniqueLocalProcess",
        "parameters": {
            "command": "echo Hello World"
        }
      }'
```

### To execute command over ssh

```bash
curl -X POST http://127.0.0.1:8080/startProcess \
  -H "Content-Type: application/json" \
  -d '{
        "name": "uniqueSSHProcess",
        "parameters": {
            "host": "192.168.1.100",
            "port": "22",
            "user": "testmest",
            "password": "superSecret",
            "value": "ls -la"
        }
      }'
```

### To copy a file to a remote machine

```bash
curl -X POST http://127.0.0.1:8080/startProcess \
  -H "Content-Type: application/json" \
  -d '{
        "name": "uniqueSCPProcess",
        "parameters": {
            "host": "192.168.1.100",
            "port": "22",
            "user": "testmest",
            "password": "superSecret",
            "localPath": "/tmp/file.txt",
            "remotePath": "/home/someuser/file.txt"
        }
      }'
```

### To list all processes

```bash
curl -X GET http://127.0.0.1:8081/listProcesses
```

### To list specific process

**Note**: You can list all processes and get an `uuid` back of a previously executed process.

```bash
curl -X GET http://127.0.0.1:8081/listProcess/123d1e08-f6d1-489a-aef6-bf782e7dc7d1
```

### To stop specific process (change process status)

```bash
curl -X POST http://127.0.0.1:8081/stopProcess/123d1e08-f6d1-489a-aef6-bf782e7dc7d1
```

### To get specific process logs

```bash
curl -X GET http://127.0.0.1:8081/processlog/123d1e08-f6d1-489a-aef6-bf782e7dc7d1
```

## Future work:
* CI/CD & Terraform for cloud deployment. Helm configuration.
* Metrics and tracing.
* Retry / back-off strategy for failing task.
* WebAPI rate-limit.
* Proper storage of secrets (Vault, k8s secrets, etc.)
* more tests...

Enjoy exploring the process app, feedback and questions are welcome!
