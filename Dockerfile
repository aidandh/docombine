FROM golang AS backend-build
WORKDIR /src
COPY backend .
RUN CGO_ENABLED=0 go build -o /bin/docombine

FROM node AS frontend-build
WORKDIR /src
COPY frontend .
RUN npm install && npm run build

FROM scratch
COPY --from=backend-build /bin/docombine .
COPY --from=frontend-build /src/build/ /static/
CMD ["./docombine"]