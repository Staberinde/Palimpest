FROM ubuntu:18.04

WORKDIR /Keep

COPY requirements.txt /Keep/requirements.txt

RUN apt update && apt install -y \
    python3.7 \
    python3.7-dev \
    python3.7-doc \
    python3-pip \
    wget

RUN pip3 install -r /Keep/requirements.txt
