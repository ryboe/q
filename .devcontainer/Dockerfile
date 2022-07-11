ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS goimage

FROM ghcr.io/ryboe/alpinecodespace:latest

COPY --from=goimage /usr/local/go/ /usr/local/go/

ENV GOPATH="/home/vscode/go"
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

# These are all the Go tools installed by the "Go: Install/Update Tools"
# command.
RUN go install github.com/haya14busa/goplay/cmd/goplay@latest
RUN go install github.com/josharian/impl@latest
RUN go install github.com/fatih/gomodifytags@latest
RUN go install github.com/cweill/gotests/gotests@latest
RUN go install mvdan.cc/gofumpt@latest
RUN go install golang.org/x/tools/gopls@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
