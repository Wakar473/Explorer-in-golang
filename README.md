# 5ire-Oracle-Service
This service is to retrieve data from CSV &amp; sign it for blockchain consensus


# How to build it

**Run by go commands**
1. `go build .`
2. `./signing_service.git`

**Run below commands with docker**
1. `docker build -t app .`
2. `docker run -p 8080:8080 -t app`

**Run using make file**
1. `make all`