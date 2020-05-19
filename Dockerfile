FROM golang:1.13

WORKDIR /app

COPY ./ /app

RUN go env -w GOPRIVATE=github.com/AJGherardi/GoMeshCryptro

RUN echo "machine github.com    login AJGherardi password 69785baa36404144f8d389e76cf79119fd632b09" > ~/.netrc 

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT ./start.sh