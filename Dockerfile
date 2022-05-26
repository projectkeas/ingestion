FROM gcr.io/distroless/static
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY ./app-${TARGETOS}-${TARGETARCH} /app/app
ENTRYPOINT [ "/app/app" ]