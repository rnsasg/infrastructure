FROM golang:latest AS builder
ARG CICD_TOKEN
ADD service /app
WORKDIR /app
# ENV GOPRIVATE=*.eng.github.com
RUN git config --global credential.helper store
#RUN echo "https://gitlab-ci-token:${CICD_TOKEN}@gitlab.eng.github.com"  >> ~/.git-credentials
RUN go mod download
RUN make build

FROM golang:latest
COPY --from=builder /app/bin/service /app/service
CMD ["/app/service"]