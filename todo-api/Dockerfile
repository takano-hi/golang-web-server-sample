FROM golang:1.19.3

WORKDIR /app/todo-api

RUN apt-get update && apt-get install -y postgresql-client
RUN go install github.com/cosmtrek/air@latest
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.1.0
RUN go install github.com/volatiletech/sqlboiler/v4@latest
RUN go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["make", "start"]
