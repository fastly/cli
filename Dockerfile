FROM rust:latest

WORKDIR /tmp
# Add some files to satisfy fastly compute build
COPY dockerfiles/* dockerfiles/.cargo ./
# Run once to force installation of some common packages
RUN fastly compute build || true
RUN rm -rf /tmp/* /tmp/.cargo

WORKDIR /app
ENTRYPOINT ["fastly"]
CMD ["--help"]
