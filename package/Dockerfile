ARG BUILD_IMAGE=registry.suse.com/bci/bci-micro:16.0
ARG BASE_IMAGE=registry.suse.com/bci/bci-nano:16.0
ARG BASE_DIGEST

FROM ${BUILD_IMAGE} AS build

ARG user=kubeapiauth

# bci-nano has no shell, prepare user files and home directory in the build
# environment
RUN mkdir -p /stage/etc /stage/home/$user && \
    echo "root:x:0:0:root:/root:/bin/bash" > /stage/etc/passwd && \
    echo "$user:x:1000:1000::/home/$user:/bin/bash" >> /stage/etc/passwd && \
    echo "root:x:0:" > /stage/etc/group && \
    echo "$user:x:1000:" >> /stage/etc/group && \
    chown -R 1000:1000 /stage/home/$user

FROM ${BASE_IMAGE}${BASE_DIGEST:+@${BASE_DIGEST}}

ARG user=kubeapiauth
ARG BASE_IMAGE
ARG BASE_DIGEST

COPY --from=build /stage/etc/passwd /stage/etc/group /etc/
COPY --from=build --chown=1000:1000 /stage/home/$user /home/$user
COPY kube-api-auth /usr/bin/

USER $user

LABEL org.opencontainers.image.base.name="${BASE_IMAGE}" \
      org.opencontainers.image.base.digest="${BASE_DIGEST}"

CMD ["kube-api-auth"]
