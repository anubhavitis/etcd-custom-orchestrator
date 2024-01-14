

# etcd-custom-orchestrator
A custom golang-library to orchestrate mutiple pods using etcd. Jobs can be allocated to pods using various allocation, along with resource security.

### Demo
[Watch demo on loom](https://www.loom.com/share/aa5d6dbf25364e6c9c1cc321d6b30496?sid=875d53b7-fabd-434a-ab3d-09ed60565492)

## Setup and Installation
> This is mac installation setup.

- Install etcd service
  ```
  $ brew install etcd
  $ brew services start etcd
  ```
- clone the repo 
   ``` 
   $ git clone https://github.com/anubhavitis/etcd-custom-orchestrator.git
   $ go mod tidy
   $ go mod vendor
   ```

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

