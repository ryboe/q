ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS goimage

FROM mcr.microsoft.com/vscode/devcontainers/base:alpine

COPY --from=goimage /usr/local/go/ /usr/local/go/

# dlv needs gcc and musl-dev. The rest are useful dev tools.
RUN apk add --update fd file fzf gcc musl-dev ripgrep

# Install the latest gh CLI tool. The first request fetches the URL for the
# latest release tarball. The second request downloads the tarball.
RUN wget --quiet --timeout=30 --output-document=- 'https://api.github.com/repos/cli/cli/releases/latest' \
    | jq -r '.assets[] | select(.name | test("gh_.*?_linux_amd64.tar.gz")).browser_download_url' \
    | wget --quiet --timeout=180 --input-file=- --output-document=- \
    | sudo tar -xvz -C /usr/local/ --strip-components=1

# Change vscode default shell to zsh.
RUN sed -i 's/\/home\/vscode:\/bin\/bash/\/home\/vscode:\/bin\/zsh/' /etc/passwd
USER vscode

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
