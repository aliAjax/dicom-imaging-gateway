build:
	go build ./...
test:
	go test ./...
race:
	go test -race ./...
vet:
	go vet ./...
fmt:
	gofmt -w cmd internal pkg
smoke:
	scripts/smoke.sh
run:
	go run ./cmd/dicom-gateway
