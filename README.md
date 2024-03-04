## Leader Election Algorithms Implementation

Final project for the course of **Distributed Systems and Cloud Computing** of the University of Rome Tor Vergata (faculty Computer Engineering), 2023/2024.
More details on the project can be found in the folder **Documentation**.

## Getting started on AWS EC2

To simulate the execution of a leader election algorithm follow the next steps:

1. Create an EC2 instance.
2. Open one terminal in the directory where you have saved the rsa key and connect to the EC2 instance:
```
ssh -i <filename>.pem ec2-user@<public-ip>
```
3. Install git on EC2 and clone this repository:
```
sudo yum update
sudo yum install git -y
git clone https://github.com/martinalupini/progetto-SDCC
```
4. Install Docker and Docker Compose:
```
./install_docker.sh
```
> [!NOTE]
> After running the script the instance will reboot. Wait a few seconds and connect again to the instance by typing from your terminal the command at point 1.

Now the environment is ready to start the simulation.

## Starting the simulation

> [!WARNING]
> This step is mandatory

At first you need to choose between "Chang-Roberts" and "Bully" by writing in the file **configuration.txt**. 
If no algorithm is selected the default algorithm is Chang-Roberts. 

- To start the simulation type from the directory **progetto-SDCC/**:
```
./Scripts/start.sh
```
After that 17 containers (15 nodes and the two registries) will be started. At the end, the node with the higher ID will be elected.

## Interacting with the simulation

In the directory **Scripts/** you can find several scripts to interact with the simulation.
Once the composition is up and running, I suggest to open a new terminal, connect to the instance and run the following scripts from the directory **progetto-SDCC/**.

- To stop all the containers type:
```
./Scripts/stop_all.sh
```
- To see more in detail how the algorithm works you can start one container at a time by typing:
```
./Scripts/start_node.sh <container name>
```
- If you want to stop one container individually (for example the leader to see how the nodes will elect a new leader) you can run:
```
./Scripts/stop_node.sh <container name>
```

> [!NOTE]
> If you want to change the alghoritm you need to build the containers' images again. To do that and also start the simulation run again:
```
docker-compose build
docker-compose up
```

## Stopping the simulation
To stop the simulation type:
```
docker-compose stop
```


