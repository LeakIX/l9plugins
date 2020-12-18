SRC= $(wildcard **/main.go)
OBJ= $(SRC:/main.go=.so)

all: $(OBJ)
%.so: %/main.go
	cd $(dir $<); go build -trimpath -ldflags "-s -w -extld=${CC}" -o ../$@ -buildmode=plugin
clean:
	rm -rf ${OBJ}