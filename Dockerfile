# Dataman-Cloud/puller docker image product release 
FROM progrium/busybox
MAINTAINER Guangzheng Zhang <zhang.elinks@gmail.com>
WORKDIR /puller
ENV PATH=$PATH:/puller/bin
RUN mkdir -p /puller/bin && \
	opkg-install ca-certificates  # prevent error complains: x509: failed to load system roots and no roots provided
COPY bin/puller /puller/bin/
ENTRYPOINT ["bin/puller", "serve"]
