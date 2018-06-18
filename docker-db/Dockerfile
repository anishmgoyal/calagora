FROM postgres

ENV POSTGRES_DB calagora
ENV POSTGRES_USER calagorauser
ENV POSTGRES_PASSWORD calagorapassword

COPY schema.sql /docker-entrypoint-initdb.d
EXPOSE 5432
