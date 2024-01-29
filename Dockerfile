FROM golang:1.21.4 as builder

RUN apt-get update && apt-get install -y ca-certificates git-core ssh tzdata
RUN mkdir -p -m 0700 /root/.ssh && ssh-keyscan github.com >> /root/.ssh/known_hosts
RUN git config --global url.ssh://git@github.com/.insteadOf https://github.com/

ARG USERID=10000
# add unprivileged user
RUN adduser --shell /bin/false --uid $USERID --disabled-login --home /app/ --no-create-home --gecos '' app \
    && sed -i -r "/^(app|root)/!d" /etc/group /etc/passwd \
    && sed -i -r 's#^(.*):[^:]*$#\1:/bin/false#' /etc/passwd

ADD . /build
WORKDIR /build

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=ssh \
    CGO_ENABLED=0 go build -o /build/cache -ldflags="-s -w" .

FROM scratch
ARG USERID=10000
# add-in our timezone data file
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# add-in our unprivileged user
COPY --from=builder /etc/passwd /etc/group /etc/shadow /etc/
# add-in our ca certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
#
COPY --from=builder --chown=$USERID:$USERID /build/cache /app/
WORKDIR /app
# from now on, run as the unprivileged user
USER $USERID
ENTRYPOINT [ "/app/cache" ]
