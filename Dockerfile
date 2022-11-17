# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.17.8-alpine as build

WORKDIR /app

ENV GO111MODULE=on
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY handler/ ./handler/ 
COPY user/ ./user/ 

ARG TARGETOS TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /hydra-id-provider

##
## Deploy
##
FROM gcr.io/distroless/base-debian11

EXPOSE 3000

USER nonroot:nonroot
WORKDIR /

COPY view/ /view/ 
COPY static/ /static/ 
COPY import/users.json /import/users.json 

COPY --from=build /hydra-id-provider /hydra-id-provider

ENTRYPOINT ["/hydra-id-provider"]
