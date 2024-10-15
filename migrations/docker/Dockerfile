FROM alpine:3.14

#COPY --from=packer /builded_app /usr/local/bin/nmshl-worker-oggr
COPY ./builded_app /usr/local/bin/migrate

ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN set -exu && \
    env && \
    chmod -R a+x /usr/local/bin/* && \
    ls -lah /usr/local/bin

COPY ./migration/database/postgres/* /usr/local/bin/migration/database/postgres/

WORKDIR /usr/local/bin/
#ENV HTTP_LISTEN_ADDRESS=":8080"

CMD ["/usr/local/bin/migrate"]

ARG BUILD_DATE
ARG LABEL_NAME
ARG LABEL_DESCRIPTION
ARG LABEL_VENDOR
ARG LABEL_URL
ARG LABEL_URL_SOURCE
ARG LABEL_URL_DOCUMENTATION
ARG LABEL_MAINTAINER

ARG CI_COMMIT_SHA
ARG CI_COMMIT_REF_NAME
ARG CI_COMMIT_MESSAGE
ARG CI_COMMIT_AUTHOR

MAINTAINER [riftbit] ErgoZ <https://riftbit.com/>

LABEL \
    org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.created="${BUILD_DATE}" \
    org.opencontainers.image.revision="${CI_COMMIT_SHA}" \
    org.opencontainers.image.title="${LABEL_NAME}" \
    org.opencontainers.image.description="${LABEL_DESCRIPTION}" \
    org.opencontainers.image.vendor="${LABEL_VENDOR}" \
    org.opencontainers.image.authors="${LABEL_MAINTAINER}" \
    org.opencontainers.image.url="${LABEL_URL}" \
    org.opencontainers.image.source="${LABEL_URL_SOURCE}" \
    org.opencontainers.image.documentation="${LABEL_URL_DOCUMENTATION}" \
	Description="${LABEL_DESCRIPTION}" \
	maintainer="${LABEL_MAINTAINER}"
