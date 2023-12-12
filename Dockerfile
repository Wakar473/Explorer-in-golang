FROM public.ecr.aws/g9h3i7k5/golang:latest as builder

# Create appuser.
# See https://stackoverflow.com/a/55757473/12429735
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /usr/src/app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go/bin/5ire-Oracle-Service

###############################################################################
# Final Stage

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
# USER appuser:appuser

ARG PACKAGE="5ire-tech/5ire-Oracle-Service"
ARG DESCRIPTION=="Default Container Image for 5ire-Oracle-Service Service"
ARG NAME="5ireChain 5ire-Oracle-Service Service"

LABEL name=${NAME} \
    maintainer="5ire Engineering <technology@5ire.org>" \
    summary=${NAME} \
    description="${DESCRIPTION}" \
    org.opencontainers.image.source="https://github.com/${PACKAGE}"\
    org.opencontainers.image.description="${DESCRIPTION}" \
    org.opencontainers.image.licenses="5ire Proprietary"

COPY --from=builder /go/bin/5ire-Oracle-Service /5ire-Oracle-Service

ENTRYPOINT ["/5ire-Oracle-Service"]