GOBUILD=go build
GOTEST=go test

# dependencies: dep_server dep_client

# dep_server:
# 	cd server && go get
# dep_client:
# 	cd client && go get

all:clean stop build_server
	./5ire-Oracle-Service 
build_server:
	$(GOBUILD) -v .
clean:
	rm -f ./5ire-Oracle-Service
stop:
	pkill 5ire-Oracle-Service || true
test:
	cd helper && $(GOTEST) -v .