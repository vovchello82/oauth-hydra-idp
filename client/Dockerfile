# syntax=docker/dockerfile:1
FROM  --platform=$BUILDPLATFORM golang:1.16-alpine as build
WORKDIR /app

ENV GO111MODULE=on
COPY go.mod ./
COPY go.sum ./
COPY client/main.go ./
RUN go mod download
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /simple-oidc-client

##
## Deploy
##
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build simple-oidc-client simple-oidc-client
COPY client/views/ /client/views/ 
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/simple-oidc-client"]