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
ENV GOTENBERG_URL=http://gotenberg:3000
ENV PORT=8080
ENV SERVE_FILES=true
CMD ["./docombine"]