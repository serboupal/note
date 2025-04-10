OUT=notec
LDFLAGS=-s -w

build:
	go build -ldflags "${LDFLAGS}" -trimpath -o ${OUT} ./cli/

run: build
	./${BINARY_NAME}

clean:
	rm ${OUT}

