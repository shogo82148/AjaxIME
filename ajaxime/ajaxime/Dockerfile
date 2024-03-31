FROM public.ecr.aws/lambda/provided:al2023 as build-image

RUN dnf install -y tar wget g++ make git python3

# install MeCab
ENV MECAB_VERSION 0.996.10
RUN mkdir -p "usr/src/mecab"
WORKDIR /usr/src/mecab
RUN wget -O mecab.tar.gz "https://github.com/shogo82148/mecab/releases/download/v${MECAB_VERSION}/mecab-${MECAB_VERSION}.tar.gz"
RUN tar -xzf mecab.tar.gz --strip-components=1
RUN ./configure --enable-utf8-only
RUN make -j$(nproc) && make install

# install MeCab dictionary
WORKDIR /usr/src
RUN git clone https://github.com/ikegami-yukino/mecab-as-kkc.git
WORKDIR /usr/src/mecab-as-kkc
RUN make
RUN MECAB_DIC_DIR=$(mecab-config --dicdir); \
    mkdir -p ${MECAB_DIC_DIR}/mecab-as-kkc; \
    install mecab-as-kkc/* ${MECAB_DIC_DIR}/mecab-as-kkc/

# install Go
COPY --from=public.ecr.aws/docker/library/golang:1.22.1-alpine3.19 /usr/local/go /usr/local/go
ENV PATH $PATH:/usr/local/go/bin

# build
WORKDIR /src
COPY go.mod go.sum main.go ./
RUN CGO_LDFLAGS=$(mecab-config --libs) CGO_CFLAGS=-I$(mecab-config --inc-dir) go build -o lambda-handler


FROM public.ecr.aws/lambda/provided:al2023

# copy the runtime executable
COPY --from=build-image /src/lambda-handler .

# copy the MeCab shared library
COPY --from=build-image /usr/local/lib/libmecab.so.* /usr/local/lib/
COPY --from=build-image /usr/local/etc/mecabrc /usr/local/etc/mecabrc
COPY --from=build-image /usr/local/lib/mecab /usr/local/lib/mecab
ENV LD_LIBRARY_PATH /usr/local/lib

RUN ./lambda-handler -version
ENTRYPOINT ./lambda-handler
