#!/bin/bash

#installing docker
sudo yum update 
sudo yum install docker 
sudo service docker start 
sudo usermod -aG docker ec2-user
sudo systemctl start docker

#installing compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

#starting the containers
docker-compose build
clear
docker-compose up
