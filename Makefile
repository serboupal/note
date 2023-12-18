OUT=notec
OUTD=noted
LDFLAGS=-s -w

build:
	go build -ldflags "${LDFLAGS}" -trimpath -o ${OUT} ./cli/
	go build -ldflags "${LDFLAGS}" -trimpath -o ${OUTD} ./rest/

run: build
	./${BINARY_NAME}

clean:
	rm ${OUT}

