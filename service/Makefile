SRC= $(wildcard **/main.go)
OBJ= $(SRC:/main.go=.so)

all: $(OBJ)
%.so: %/main.go
	cd $(dir $<); go build -trimpath -ldflags "-s -w" -o ../$@ -buildmode=plugin
clean:
	rm -rf ${OBJ}