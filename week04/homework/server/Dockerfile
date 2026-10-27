FROM golang:1.25-bookworm AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /study-agent-api .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends poppler-utils tesseract-ocr tesseract-ocr-chi-sim tesseract-ocr-eng ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /study-agent-api /usr/local/bin/study-agent-api
ENV TESSERACT_PATH=/usr/bin/tesseract
EXPOSE 8080
CMD ["study-agent-api"]
