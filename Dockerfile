FROM golang:alpine as base

RUN mkdir /url-shortener
WORKDIR /url-shortener
COPY go.mod .
COPY go.sum .

RUN go mod download
# COPY the source code as the last step
COPY . .

FROM base as development
# Go fresh detects changes in the source code and re-runs the application
RUN go get github.com/pilu/fresh
CMD ["fresh"]

# Build the binary
FROM base as pre-build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/url-shortener
FROM scratch as production
COPY --from=pre-build /go/bin/url-shortener /
COPY --from=base /url-shortener/static /static
#COPY --from=pre-build /go/bin/url-shortener /go/bin/url-shortener
#COPY --from=base /url-shortener/static /go/bin/static

ENTRYPOINT ["/url-shortener"]
