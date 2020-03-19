FROM golang:1.13

WORKDIR /app

COPY ./ /app

COPY .netrc ~/.netrc

RUN go env -w GOPRIVATE=github.com/AJGherardi/GoMeshCryptro

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon --build="go build ." --command=./HomeHub