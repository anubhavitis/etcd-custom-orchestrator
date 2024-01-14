# etcd-custom-orchestrator
A custom golang-library to orchestrate mutiple pods using etcd. Jobs can be allocated to pods using various allocation, along with resource security.

### Demo

<div style="position: relative; padding-bottom: 64.98194945848375%; height: 0;"><iframe src="https://www.loom.com/embed/aa5d6dbf25364e6c9c1cc321d6b30496?sid=1c8e0ab0-2331-48dd-a093-5beff3eba7f7" frameborder="0" webkitallowfullscreen mozallowfullscreen allowfullscreen style="position: absolute; top: 0; left: 0; width: 100%; height: 100%;"></iframe></div>

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

