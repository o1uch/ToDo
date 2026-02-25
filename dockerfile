FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o scheduler ./main.go

FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /app/scheduler .
COPY --from=builder /app/web ./web

ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db
ENV TODO_PASSWORD=12345  

EXPOSE ${TODO_PORT}

CMD ["./scheduler"]

# команда для запуска контейнера: docker run -d --name scheduler-container -p 7540:7540 -v "C:\path\to\db\scheduler.db:/data/scheduler.db" -e TODO_PASSWORD=12345 my-scheduler