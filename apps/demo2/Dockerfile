FROM golang

COPY call .
RUN chmod +x call
ENV REGION=ap-northeast-2
ENV CLOUDMAP_SERVICE_NAME=ecs
CMD ["./call"]
