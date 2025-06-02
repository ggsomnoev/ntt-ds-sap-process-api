# CLI Process Manager Requirements

## Overview

CLI application written in Go that launches an HTTP RESTful web server and handle process commands. Each process consists of multiple tasks that can be executed locally, over SSH, or through SCP.

## HTTP API Endpoints

The web server should expose the following endpoints:

- `POST /startProcess`  
  Starts a new process. Expects a JSON payload containing the startup configuration.

- `GET /listProcesses`  
  Returns a list of all running processes.

- `GET /listProcess/{ID}`  
  Returns details for a specific process by its ID.

- `POST /stopProcess/{ID}`  
  Stops the process identified by the given ID.

- `GET /processlog/{ID}`  
  Retrieves the log file associated with a specific process.

## Process Definition (YAML)

Processes must be defined in a YAML file. The structure is as follows:

### Structure

```yaml
name: uniqueProcessName

params:
  - name: parameterName
    mandatory: true
    description: description of the parameter
    defvalue: ""

tasks:
  - name: sshTask
    class: sshCmd
    parameters:
      parm1: "{{.value}}"  

  - name: parallel_local_task
    class: localCmd
    parameters:
      command: "{{.command}}"  

  - name: sync_local_task
    class: localCmd
    parameters:
      command: "{{.command}}"
    waitfor:
      - parallel_local_task
```

### Notes

- **Process Name** must be unique. Duplicate names are not allowed.
- **Parameters** declared in the `params` section should be available for task templating using Go’s templating engine.
- **Task Types** must include:
  - `localCmd` – executes a command in the local terminal.
  - `sshCmd` – executes a command over SSH on a remote machine.
  - `scpCmd` – copies a file to a remote host.
- **Parallel Execution** of tasks must be supported.
- **Task Synchronization** can be controlled using the `waitfor` field.
