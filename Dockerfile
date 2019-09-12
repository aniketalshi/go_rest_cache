#  golang 1.13 base image
FROM golang:1.12-alpine

# ARGS for env variables
ARG GITHUB_API_TOKEN
ARG REDIS_URL

# These env variables would come from docker-compose file
ENV GITHUB_API_TOKEN $GITHUB_API_TOKEN
ENV REDIS_URL $REDIS_URL

# Alpine images tends to not have git and bash tools
RUN apk update && apk upgrade && apk add --no-cache bash git openssh

LABEL maintainer="Aniket Alshi <aniketalshi@gmail.com>"

WORKDIR /app
COPY go.mod go.sum ./

# Download and cache all dependencies
RUN go mod download

# copy source from current directory to working directory
COPY . .

# Builds
RUN go build -o main .

EXPOSE 3000

# Run the executable
CMD ["./main"]
