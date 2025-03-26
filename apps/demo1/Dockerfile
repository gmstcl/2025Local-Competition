FROM golang

COPY save . 
RUN chmod +x save 
ENV REGION=ap-northeast-2
ENV DYNAMODB_TABLE=dev
ENV CLOUDMAP_SERVICE_NAME=ecs
CMD ["./save"]
