# ---------- 构建阶段：Vite 打包静态产物 ----------
FROM node:20-alpine AS build
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm config set registry https://registry.npmmirror.com && npm install

COPY . .
RUN npm run build

# ---------- 运行阶段：Nginx 托管 SPA + 反代 ----------
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
