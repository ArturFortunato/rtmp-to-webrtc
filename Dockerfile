FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main *.go
CMD ["/app/main"]

EXPOSE 1935 8080
