FROM golang:alpine

RUN mkdir /url-shortener
WORKDIR /url-shortener
COPY go.mod .
COPY go.sum .

# Get dependancies
RUN go mod download
# COPY the source code as the last step
COPY . .

# Go fresh detects changes in the source code and re-runs the application
RUN go get github.com/pilu/fresh
CMD ["fresh"]