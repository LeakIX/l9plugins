all: goget build-service
goget:
	go get -d ./...
build-service:
	$(MAKE) -C service
clean:
	$(MAKE) -C service clean
.PHONY: all $(SUBDIRS)