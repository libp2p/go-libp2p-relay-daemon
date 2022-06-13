FROM  golang:alpine as builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./
RUN ls
RUN go build -o /libp2p-relay-daemon ./libp2p-relay-daemon


FROM alpine
COPY --from=builder /libp2p-relay-daemon ./
EXPOSE 6000/tcp
EXPOSE 4001/tcp
EXPOSE 4001/udp
ENTRYPOINT [ "./libp2p-relay-daemon" ]
