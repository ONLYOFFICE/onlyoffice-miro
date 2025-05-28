FROM node:20-alpine AS builder
WORKDIR /app

ARG VITE_MIRO_ONLYOFFICE_BACKEND
ARG VITE_MIRO_INSTALLATION_URL
ARG VITE_ASC_DEMO_EXPIRATION_DAYS

ENV VITE_MIRO_ONLYOFFICE_BACKEND=${VITE_MIRO_ONLYOFFICE_BACKEND}
ENV VITE_MIRO_INSTALLATION_URL=${VITE_MIRO_INSTALLATION_URL}
ENV VITE_ASC_DEMO_EXPIRATION_DAYS=${VITE_ASC_DEMO_EXPIRATION_DAYS}

COPY package*.json ./
COPY tsconfig.json ./
COPY vite.config.mts ./
COPY index.html ./
COPY application.html ./
COPY .eslintrc.cjs ./

RUN npm install
COPY src ./src
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"] 
