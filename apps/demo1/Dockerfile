FROM golang

COPY main . 
RUN chmod +x main 
ENV REGION=ap-northeast-2
ENV DYNAMODB_TABLE=dev
ENV CLOUDMAP_SERVICE_NAME=ecs
CMD ["./main"]
