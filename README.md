# etcd-custom-orchestrator
A custom golang-library to orchestrate mutiple pods using etcd. Jobs can be allocated to pods using various allocation, along with resource security.

## Setup and Installation
> This is mac installation setup.

- Install etcd service
  - ```$ brew install etcd```
  - ```$ brew services start etcd```
- clone the repo 
  - ``` $ git clone https://github.com/anubhavitis/etcd-custom-orchestrator.git ```
  - ```$ go mod tidy```
  - ```$ go mod vendor```

### Steps to Run

- Run on two different ports:
  ```
  $ go run main.go -port :8080
  ```
  ```
  $ go run main.go -port :8081
  ```
- Now use another terminal to add new jobs:
  ```
  $ curl http://localhost:8080/<job-name>
  ```

### Allocation strategy

Currently we are only using hash-based allocation, but can be extended to weight based allocation, and other strategies as well.

