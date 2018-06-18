FROM caldeps

# All folders related to the web app
ADD css/ /cl/css
ADD img/ /cl/img
ADD js/ /cl/js
ADD views/ /cl/views

ADD . /cl/src/go/src/github.com/anishmgoyal/calagora

ENV GOPATH=/cl/src/go
ENV CALAGORA_DB_HOST "db:5432"
ENV CALAGORA_SAVE_DIR "/cl/files/"

# Build Calagora
WORKDIR /cl/src/go/src/github.com/anishmgoyal/calagora
RUN go build && mv calagora /cl/calagora

WORKDIR /cl
RUN rm -rf /cl/src

# Ensure we have a folder for local uploads and persistent uploads
RUN mkdir -p /cl/files
RUN mkdir -p /cl/tmp

EXPOSE 2646

ENTRYPOINT [ "/cl/calagora" ]