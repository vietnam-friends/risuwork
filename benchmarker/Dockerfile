FROM golang:1.22-bookworm AS base

WORKDIR /bench

RUN go env -w GOMODCACHE=/tmp/go/.cache
RUN --mount=type=cache,uid=1001,target=/tmp/go/.cache \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

FROM base AS build

WORKDIR /bench

RUN go env -w GOMODCACHE=/tmp/go/.cache
RUN go env -w GOCACHE=/tmp/go/.cache
RUN --mount=type=cache,uid=1001,target=/tmp/go/.cache \
    --mount=type=bind,source=.,target=. \
    CGO_ENABLED=0 go build -o /bin/bench ./cmd/bench/...

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/bench /bin/

ENV QUEUE_URL="https://sqs.ap-northeast-1.amazonaws.com/${ACCOUNT_ID}/bench-trigger-queue.fifo"
ENV TOPIC_ARN="arn:aws:sns:ap-northeast-1:${ACCOUNT_ID}:bench-score-topic"
ENV TABLE_NAME="benchmark_results"

CMD ["/bin/bench", "supervise"]
