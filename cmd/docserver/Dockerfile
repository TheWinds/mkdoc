FROM alpine:3.7
COPY ./mkdoc /mkdoc/mkdoc
COPY ./docserver /mkdoc/server
WORKDIR /mkdoc
ENV PATH="/mkdoc:${PATH}"
ENV GOPATH="/mkdoc"
CMD "./server"