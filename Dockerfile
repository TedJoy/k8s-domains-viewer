FROM golang:1.18 AS build

WORKDIR /app
COPY . /app
RUN go mod tidy \
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o out .

FROM alpine
WORKDIR /app
RUN apk --no-cache add aws-cli tini skopeo
COPY --from=build /app/out /app/out
ENTRYPOINT [ "tini", "--" ]
CMD [ "/app/out" ]
