FROM golang:1.25-alpine AS build

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o benchmark ./main.go

FROM scratch AS final
COPY --from=build /build/benchmark /benchmark
CMD [ "/myapp" ]