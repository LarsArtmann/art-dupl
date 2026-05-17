FROM gcr.io/distroless/static-debian13:nonroot

COPY art-dupl /art-dupl

USER 65532:65532

ENTRYPOINT ["/art-dupl"]
