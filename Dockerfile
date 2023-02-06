FROM golang:1.18 as builder
WORKDIR /build
COPY . ./
RUN go build -o /libp2p-relay-daemon ./cmd/libp2p-relay-daemon


FROM alpine
COPY --from=builder /libp2p-relay-daemon ./
EXPOSE 6000/tcp
EXPOSE 4001/tcp
EXPOSE 4001/udp
ENTRYPOINT [ "./libp2p-relay-daemon" ]
