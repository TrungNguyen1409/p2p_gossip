docker build -f Dockerfile.bootstrapper -t my-bootstrapper .
docker build -t my-node-image-new-config .
docker compose up -d