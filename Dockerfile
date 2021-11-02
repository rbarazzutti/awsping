FROM golang:1.17-bullseye as build
WORKDIR /build
COPY go.mod go.sum /build/
RUN go mod download
COPY . /build
RUN go build

FROM gcr.io/distroless/base
COPY --from=build /build/awsping /
ENTRYPOINT ["/awsping"]
