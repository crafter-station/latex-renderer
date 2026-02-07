FROM public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 AS adapter
FROM golang:1.25 AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

# ---- Runtime image ----
FROM ubuntu:22.04

# Lambda Web Adapter: translates Lambda events to HTTP requests
COPY --from=adapter /lambda-adapter /opt/extensions/lambda-adapter

ENV DEBIAN_FRONTEND=noninteractive
ENV AWS_LWA_PORT=8080

RUN apt update && apt install -y \
  build-essential \
  perl \
  cpanminus \
  ca-certificates \
  libxml2-dev \
  libxslt1-dev \
  zlib1g-dev \
  pkg-config \
  && rm -rf /var/lib/apt/lists/*

RUN apt update && apt install -y \
  texlive \
  texlive-latex-recommended \
  texlive-latex-extra \
  texlive-fonts-recommended \
  && rm -rf /var/lib/apt/lists/*

ENV C_INCLUDE_PATH=/usr/include/libxml2
RUN cpanm --verbose XML::LibXML
RUN cpanm --verbose XML::LibXSLT
RUN cpanm --verbose JSON::XS
RUN cpanm --notest LaTeXML

WORKDIR /app
COPY --from=build /app/server .

EXPOSE 8080
CMD ["./server"]
