FROM golang AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/docombine

FROM scratch
COPY --from=build /bin/docombine .
CMD ["./docombine"]