FROM ryboe/archcodespace:latest

ENV GOPATH="/home/vscode/go"
ENV PATH="$GOPATH/bin:$PATH"

RUN sudo pacman -Syu --noconfirm go

RUN go install github.com/cweill/gotests/gotests@latest \
    && go install github.com/fatih/gomodifytags@latest \
    && go install github.com/josharian/impl@latest \
    && go install github.com/haya14busa/goplay/cmd/goplay@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install golang.org/x/tools/gopls@latest \
    && go install mvdan.cc/gofumpt@latest
