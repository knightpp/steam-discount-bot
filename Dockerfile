FROM golang:alpine as build
WORKDIR /app
COPY . .
# RUN go get -d -v ./...
RUN go build \
	-o steam-discount \
	.

FROM alpine
WORKDIR /app

COPY --from=build /app/steam-discount .
COPY --from=build /app/.env .

CMD [ "/app/steam-discount" ]
