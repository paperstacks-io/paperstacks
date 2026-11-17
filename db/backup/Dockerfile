FROM postgres:18-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends barman rsync \
    && rm -rf /var/lib/apt/lists/* \
    && chown postgres:postgres /var/lib/barman

COPY barman.conf /etc/paperstacks/barman.conf
COPY entrypoint.sh /usr/local/bin/paperstacks-barman
RUN chmod 755 /usr/local/bin/paperstacks-barman

USER postgres
ENTRYPOINT ["/usr/local/bin/paperstacks-barman"]
