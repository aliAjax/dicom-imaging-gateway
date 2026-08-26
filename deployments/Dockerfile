FROM golang:1.26.6 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/dicom-gateway ./cmd/dicom-gateway
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/dicom-gateway /dicom-gateway
EXPOSE 8080
ENTRYPOINT ["/dicom-gateway"]
