FROM gcr.io/distroless/static
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY ./ingestion-${TARGETOS}-${TARGETARCH} /app/ingestion
ENTRYPOINT [ "/app/ingestion" ]