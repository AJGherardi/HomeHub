FROM golang:1.15.2-buster AS build

WORKDIR /app

COPY ./ /app

RUN apt update

RUN apt install -y libusb-1.0-0-dev

RUN go build -o /bin/home

FROM golang:1.15.2-buster

RUN apt update

RUN apt install -y libusb-1.0

COPY --from=build /bin/home /bin/home

ENTRYPOINT ["/bin/home"]