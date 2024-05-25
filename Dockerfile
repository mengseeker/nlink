FROM ubuntu

RUN apt update && apt install -y vim curl

WORKDIR /app

ADD build/nlink /usr/local/bin/nlink
ADD nlink.yaml nlink.example.yaml

CMD ["bash"]