FROM deathstarbench/media-microservices-deps:latest

COPY ./ /media-microservices
RUN cd /media-microservices \
    && mkdir -p build \
    && cd build \
    && cmake .. \
    && make \
    && make install

WORKDIR /media-microservices