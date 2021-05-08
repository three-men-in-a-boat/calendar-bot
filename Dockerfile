# Step 1. build_step
FROM golang:1.16-buster AS builder
ARG APP=/app
WORKDIR ${APP}
COPY go.mod .
RUN go mod download
COPY cmd cmd
COPY pkg pkg
COPY Makefile .
RUN make inside-docker-build

# Step 2. release_step
FROM debian:buster-slim AS main
ARG APP=/app
ENV TZ=Etc/UTC \
    APP_USER=calendar-bot
RUN groupadd $APP_USER \
    && useradd -r -g $APP_USER $APP_USER
USER $APP_USER
EXPOSE 2000
EXPOSE 8080
EXPOSE 8081
WORKDIR ${APP}
COPY --from=builder /app/build/bin/botbackend botbackend
CMD ["./botbackend"]
