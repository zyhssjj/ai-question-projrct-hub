# 基于官方Nginx镜像（alpine版体积小）
FROM nginx:alpine

# 删除Nginx默认配置（避免冲突）
RUN rm /etc/nginx/conf.d/default.conf

# 将本地的nginx.conf复制到容器的Nginx配置目录
COPY nginx.conf /etc/nginx/conf.d/

# 将本地dist目录的前端静态资源复制到Nginx默认静态文件目录
COPY dist/ /usr/share/nginx/html/

# 暴露容器的80端口（Nginx默认端口）
EXPOSE 80

# 启动Nginx（容器启动时自动执行）
CMD ["nginx", "-g", "daemon off;"]