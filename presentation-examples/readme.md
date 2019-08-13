# Golang Examples

This is a quick introduction of Golang. You can use these scripts to explain basic functionalities of Golang

- The example 1 is the basic "Hello world"
    go run ./example1/hello.go
You can also compile and run 
    go build ./example1/hello.go
    go run hello //Linux
    go run hello.exe //Windows
- The example 2 is a basic "Hello World" with a web node.
    go run ./example2/helloweb.go
- The example 3 is a basic explanation (with a lot of comments) of the power of Go.
    go run ./example3/helloweb-args.go //Error verification
    go run ./example3/helloweb-args.go 8081
    go run ./example3/helloweb-args.go 80810 //Error verification
- The example 4 (or advanced) is a explanation of init, regexp, flags, file & packages. All the great things that makes Go great.
    go run ./example4/goInRealLife.go
    go run ./example4/goInRealLife.go -o "out.json"