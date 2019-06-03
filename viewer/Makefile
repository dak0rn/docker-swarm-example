# Runs the application using `realize`
# Requires `realize` to be installed globally
run: main.go
	modd

# Builds the application
# Strips debugging information and path leaks
build: main.go
	CGO_ENABLED=0 go build \
	-a -o viewer \
	-gcflags "all=-trimpath=${CURDIR}" \
	-asmflags "all=-trimpath=${CURDIR}" \
	*.go
