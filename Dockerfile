from golang:1.19

EXPOSE 80

ENV RESOURCE_PATH=/app/resource
ENV VOLUME_PATH=/app/volume

COPY . /app
WORKDIR /app

RUN ["go", "get", "github.com/mattn/go-sqlite3"]
RUN ["go", "get", "github.com/go-chi/chi/v5"]
RUN ["go", "get", "golang.org/x/net/html"]

RUN ["go", "build", "-o", "executable", "cmd/main.go"]

ENTRYPOINT ["/app/executable"]
