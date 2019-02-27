FROM node:11.2.0-slim

RUN mkdir -p /src/app
WORKDIR /src/app
# Install app dependencies
# A wildcard is used to ensure both package.json AND package-lock.json are copied
# where available (npm@5+)
COPY package*.json ./
COPY customerlist.json ./
COPY server.js ./

RUN npm install
# If you are building your code for production
# RUN npm ci --only=production


EXPOSE 80
CMD node server.js
