FROM scratch
COPY apienroll /
ENV PORT 9000
ENTRYPOINT ["/apienroll"]
