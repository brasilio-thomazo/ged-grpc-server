ARG UID=1000
ARG GID=1000
####################################################################################
#                                      ALPINE                                      #
#                                       BASE                                       #
####################################################################################
FROM alpine:latest as base
ARG GID
ARG UID
RUN apk add --no-cache bash curl doas \
    && addgroup -g ${GID} app \
    && adduser -G wheel -u ${UID} -s /bin/bash -D app app \
    && mkdir -p /home/app/public_html \
    && chown app:app /home/app -R \
    && echo 'permit nopass :wheel as root' >> /etc/doas.d/doas.conf

####################################################################################
#                                   BUILD GRPC                                     #
#                                     SERVER                                       #
####################################################################################
FROM base as build
COPY . /source
WORKDIR /source
RUN apk add --no-cache go \
    && go mod tidy \
    && go build -v -o hermes .

####################################################################################
#                                      GRPC                                        #
#                                     SERVER                                       #
####################################################################################
FROM base as server
ENV GRPC_PORT=50051
ENV DB_WRITER_HOST=localhost
ENV DB_WRITER_PORT=5432
ENV DB_READER_HOST=localhost
ENV DB_READER_PORT=5432
ENV DB_USERNAME=postgres
ENV DB_PASSWORD=
ENV DB_DATABASE=postgres
ENV APP_PATH=
COPY --from=build /source/hermes /usr/local/bin/
USER app
WORKDIR /home/app/public_html/storage/app
CMD [ "hermes" ]