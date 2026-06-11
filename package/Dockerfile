ARG BASE_IMAGE=registry.suse.com/bci/bci-micro:15.7
ARG BASE_DIGEST

FROM ${BASE_IMAGE}${BASE_DIGEST:+@${BASE_DIGEST}}

ARG user=kubeapiauth
ARG BASE_IMAGE
ARG BASE_DIGEST

RUN echo "$user:x:1000:1000::/home/$user:/bin/bash" >> /etc/passwd && \
    echo "$user:x:1000:" >> /etc/group && \
    mkdir /home/$user && \
    chown -R $user:$user /home/$user

COPY kube-api-auth /usr/bin/

USER $user

LABEL org.opencontainers.image.base.name="${BASE_IMAGE}" \
      org.opencontainers.image.base.digest="${BASE_DIGEST}"

CMD ["kube-api-auth", "serve"]
