FROM golang:1.14.3-alpine3.11 AS BUILD

RUN apk add build-base

WORKDIR /app
ADD /go.mod /app
ADD /go.sum /app
RUN go mod download

#now build source code
ADD / /app
RUN go build -o /go/bin/userme-demo-api


FROM golang:1.14.3-alpine3.11

ENV LOG_LEVEL                           'info'
ENV CORS_ALLOWED_ORIGINS                '*'
ENV JWT_SIGNING_METHOD                  'ES256'
ENV JWT_SIGNING_KEY_FILE                '/run/secrets/jwt-verify-key'
ENV BASE_SERVER_URL_FOR_LOCATIONS       ''

COPY --from=BUILD /go/bin/userme-demo-api /bin/

ADD startup.sh /

EXPOSE 6000

CMD [ "/startup.sh" ]

