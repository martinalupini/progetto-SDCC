#!/bin/bash

sudo service docker start
docker-compose build 
clear
docker-compose up
