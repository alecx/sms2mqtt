# The service binary is pre-built (pure Go, cross-compiled) and shipped per-arch,
# so the on-device build is just a COPY — no Go toolchain in the image.
ARG BUILD_FROM
FROM ${BUILD_FROM}

# BUILD_ARCH is supplied by the Supervisor (aarch64 / amd64) and matches the
# binary suffixes produced by build.sh.
ARG BUILD_ARCH

COPY run.sh /
COPY bin/sms2mqtt.${BUILD_ARCH} /usr/bin/sms2mqtt
RUN chmod a+x /run.sh /usr/bin/sms2mqtt

CMD [ "/run.sh" ]
