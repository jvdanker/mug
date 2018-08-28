FROM gcc

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    libfreeimage3 \
    libfreeimage-dev

RUN wget -O /tmp/pdiff.tar.gz 'https://excellmedia.dl.sourceforge.net/project/pdiff/pdiff/perceptualdiff-1.1.1/perceptualdiff-1.1.1-linux.tar.gz'; \
    mkdir /pdiff; \
    tar -xf /tmp/pdiff.tar.gz -C /pdiff; \
    ls -l /pdiff; \
    cmake .; \
    make .

WORKDIR /images

ENTRYPOINT ["/pdiff/perceptualdiff"]