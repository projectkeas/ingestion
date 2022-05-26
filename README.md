# Service Template

## Checklist

- Replace `appName` in `./.github/workflows/PR-Build.yml`
- Replace `appName` in `./.github/workflows/Release-Published.yml`
- Replace `appName` in `./app.go`
- Replace `app` in `./Dockerfile` (_see below_)

## Dockerfile changes

If the app you are building is going to be called `connector`, the `Dockerfile` needs to change from:

```Dockerfile
FROM gcr.io/distroless/static
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY ./app-${TARGETOS}-${TARGETARCH} /app/app
ENTRYPOINT [ "/app/app" ]
```

to:

```Dockerfile
FROM gcr.io/distroless/static
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY ./connector-${TARGETOS}-${TARGETARCH} /app/connector
ENTRYPOINT [ "/app/connector" ]
```
