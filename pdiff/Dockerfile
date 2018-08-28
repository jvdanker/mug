FROM gcc

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    libfreeimage3 \
    libfreeimage-dev

ADD src/* /pdiff/

RUN cd; cd /pdiff; cmake .; make ; ls -lrtR /pdiff

WORKDIR /images

ENTRYPOINT ["/pdiff/perceptualdiff"]