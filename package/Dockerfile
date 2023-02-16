FROM registry.suse.com/bci/bci-micro:15.4.17.1

ARG user=kubeapiauth

RUN echo "$user:x:1000:1000::/home/$user:/bin/bash" >> /etc/passwd && \
    echo "$user:x:1000:" >> /etc/group && \
    mkdir /home/$user && \
    chown -R $user:$user /home/$user

COPY kube-api-auth /usr/bin/

USER $user

CMD kube-api-auth serve
