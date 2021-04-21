# docker build -t fastly/cli . -f ./Dockerfile
# docker run -it fastly/cli

FROM debian:stable-slim

WORKDIR /tmp

RUN apt-get update && apt-get install -y curl jq

RUN export FASTLY_CLI_VERSION=$(curl --silent https://api.github.com/repos/fastly/cli/releases/latest | jq -r .tag_name | cut -d 'v' -f 2) \
  GOARCH=$(dpkg --print-architecture) \
  && curl -vL "https://github.com/fastly/cli/releases/download/v${FASTLY_CLI_VERSION}/fastly_v${FASTLY_CLI_VERSION}_linux-$GOARCH.tar.gz" -o fastly.tar.gz
RUN tar -xzf fastly.tar.gz
RUN mv fastly /usr/bin

WORKDIR /app
ENTRYPOINT ["/usr/bin/fastly"]
CMD ["--help"]
