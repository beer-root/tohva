depends:
	go get -v -d

all: depends
	go build tohva.go
