FROM ubuntu:24.04

ARG TARGETOS
ARG TARGETARCH
ARG ENV

ARG DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN echo "Building for platform: ${TARGETOS} / ${TARGETARCH}"

RUN apt-get update \
    && apt-get install -y ca-certificates vim telnet \
    && rm -rf /var/lib/apt/lists/* 
WORKDIR /opt/heero

COPY bin/[[.ServiceName]]_${TARGETOS}_${TARGETARCH} [[.ServiceName]]
CMD ["./[[.ServiceName]]"]