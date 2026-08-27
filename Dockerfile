# syntax=docker/dockerfile:1.7
FROM golang:1.27.0-alpine3.23 AS go-build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/http-repro ./cmd/http-repro \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mock-api ./cmd/mock-api

FROM gcr.io/distroless/static-debian12:nonroot AS cli
COPY --from=go-build /out/http-repro /usr/local/bin/http-repro
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/http-repro"]

FROM gcr.io/distroless/static-debian12:nonroot AS mock-api
COPY --from=go-build /out/mock-api /usr/local/bin/mock-api
USER nonroot:nonroot
EXPOSE 9090
ENV MOCK_API_ADDR=0.0.0.0:9090
ENTRYPOINT ["/usr/local/bin/mock-api"]

FROM go-build AS report-build
COPY fixtures ./fixtures
RUN /out/http-repro analyze fixtures/auth-401.har --output /report

FROM nginxinc/nginx-unprivileged:1.29.3-alpine3.22 AS web
COPY --from=report-build /report /usr/share/nginx/html
USER 101:101
EXPOSE 8080

