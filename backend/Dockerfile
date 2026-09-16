FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /bootstrap ./cmd/lambda

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /bootstrap ${LAMBDA_TASK_ROOT}/bootstrap
CMD ["bootstrap"]