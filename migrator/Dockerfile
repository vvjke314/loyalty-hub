FROM golang:1.24-bullseye

RUN apt-get update && apt-get install -y git 
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /app
COPY ./migrations /migrations
COPY .env .env

ENV PATH="/go/bin:$PATH"

ENTRYPOINT ["goose", "-dir", "/migrations", "postgres"]
CMD ["up"]