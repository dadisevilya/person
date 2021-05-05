FROM golang:1.11.4-alpine3.8 as base_builder

RUN apk add --update alpine-sdk gcc git vim bash && rm -rf /var/cache/apk/*

ARG BUNDLE_GITHUB__COM
RUN git config --global url."https://$BUNDLE_GITHUB__COM:x-oauth-basic@github.com/".insteadOf "https://github.com/"

ARG APP_NAME
ENV APP_NAME $APP_NAME
ADD . /go/src/github.com/gtforge/$APP_NAME
WORKDIR /go/src/github.com/gtforge/$APP_NAME

# Force the go compiler to use modules
ENV GO111MODULE=on

# We want to populate the module cache based on the go.{mod,sum} files.
RUN go mod download

RUN go get github.com/gtforge/swan

FROM base_builder AS app_builder
RUN GOGC=off go build -i -v -ldflags "-X main.Buildstamp=`date -u +%Y/%m/%d_%H:%M:%S` -X main.Commit=`git rev-parse HEAD`" -o backend cmd/backend/*.go

FROM alpine:latest
RUN apk add --update tzdata ca-certificates bash

RUN echo 'alias ll="ls -la"' >> ~/.bashrc

WORKDIR /app/
ARG APP_NAME
COPY --from=app_builder /go/src/github.com/gtforge/$APP_NAME/scripts/migrate /bin/migrate
COPY --from=app_builder /go/bin/swan /usr/bin/
COPY --from=app_builder /go/src/github.com/gtforge/$APP_NAME/backend .
COPY --from=app_builder /go/src/github.com/gtforge/$APP_NAME/config config
COPY --from=app_builder /go/src/github.com/gtforge/$APP_NAME/migrations migrations
ENV APP_NAME $APP_NAME
ENV SERVICE_NAME $APP_NAME
CMD ["./backend"]
EXPOSE 80
