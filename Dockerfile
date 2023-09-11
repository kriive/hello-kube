FROM golang:1.21.1-alpine3.18 as build

# Change directory in work.
WORKDIR /work

# Copy go.mod and go.sum to allow caching.
COPY go.mod .
COPY go.sum .

RUN go mod download

# Copy source files.
COPY *.go .
COPY http ./http
COPY cmd ./cmd

ENV VERSION="0.0.0-development"
ENV COMMIT="deadbeef"

RUN CGO_ENABLED=0 go build -ldflags "-X 'main.version=$VERSION' -X 'main.commit=$COMMIT'" -o hellod ./cmd/hellod

FROM gcr.io/distroless/static-debian11:nonroot

COPY --from=build /work/hellod /

CMD ["/hellod"]