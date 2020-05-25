FROM golang:alpine as build-env


# WORKDIR /url-shortener
# COPY . .

# RUN go get -d -v ./...
# RUN go build

# EXPOSE 9999

# CMD ["url-shortener"]

# All these steps will be cached
RUN mkdir /url-shortener
WORKDIR /url-shortener
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/url-shortener
FROM scratch
COPY --from=build-env /go/bin/url-shortener /go/bin/url-shortener

EXPOSE 9999
CMD ["/go/bin/url-shortener"]

VOLUME "/go/stc"

# ENTRYPOINT ["/go/bin/url-shortener"]