FROM node AS frontend-build
WORKDIR /src
COPY frontend .
RUN npm install && npm run build

FROM nginx
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=frontend-build /src/build/ /usr/share/nginx/html/
CMD ["nginx", "-g", "daemon off;"]