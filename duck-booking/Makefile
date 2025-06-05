APP_EXECUTABLE=ducks

build:
	GOARCH=amd64 go build -o ${APP_EXECUTABLE} main.go

run: build
	./${APP_EXECUTABLE}

clean:
	 go clean
